import MySQLdb
import string
import random
import pymongo
import sys

is_sql = sys.argv[1] == "mysql";

if is_sql:
	db = MySQLdb.connect(host="localhost", user="root", passwd="root", db="statstat", charset='utf8')
else:
	connection = pymongo.Connection()
	db = connection.statstat
	res = db.stats

projects = [1,2,3,4]
types = [20,21,22,23,24,25,26,27]
meta_types = [1030, 1031, 1032]

for i in range(0,10000000):
	stat = { "project_id":random.choice(projects), "type_id":random.choice(types), "value":random.randint(1,10000) }
	if is_sql:
		sql = """INSERT INTO stats(project_id, type_id, value) VALUES ('%(project_id)s', '%(type_id)s', '%(value)s');"""%stat
	else:
		res.insert(stat)

	for j in range(1,3):
		stat_meta = { "project_id":random.choice(projects), "type_id":random.choice(types), "value":random.randint(1,10000) , "meta_type_id": random.choice(meta_types), "meta_value": random.randint(1,30000) }
		if is_sql:
			sql = sql+"""INSERT INTO stats(project_id, type_id, value, meta_type_id, meta_value) VALUES ('%(project_id)s', '%(type_id)s', '%(value)s', '%(meta_type_id)s', '%(meta_value)s');"""%stat_meta
		else:
			res.insert(stat_meta)

	if is_sql:
		res = db.cursor()
		res.execute(sql)

	if i % 10000 == 0:
		print i

db.close()
