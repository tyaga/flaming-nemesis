import MySQLdb
import string
import random

db = MySQLdb.connect(host="localhost", user="root", passwd="root", db="statstat", charset='utf8')
cursor = db.cursor()

projects = [1,2,3,4]
types = [20,21,22,23,24,25,26,27]
meta_types = [1030, 1031, 1032]

for i in range(0,10000000):

	project_id = random.choice(projects)
	type_id = random.choice(types)
	value = random.randint(1,10000)

	sql = """INSERT INTO stats(project_id, type_id, value) VALUES ('%(project_id)s', '%(type_id)s', '%(value)s');
	"""%{"project_id":project_id, "type_id":type_id, "value":value}
	print sql
#	cursor.execute(sql)

	for j in range(1,3):
		meta_type_id = random.choice(meta_types)
		meta_value = random.randint(1,30000)
		sql = """INSERT INTO stats(project_id, type_id, value, meta_type_id, meta_value) VALUES ('%(project_id)s', '%(type_id)s', '%(value)s', '%(meta_type_id)s', '%(meta_value)s');
		"""%{"project_id":project_id, "type_id":type_id, "value":value, "meta_type_id":meta_type_id, "meta_value": meta_value}
		print sql
#		cursor.execute(sql)

	if i % 10000 == 0: 
		#print i
		db.commit()

db.close()
