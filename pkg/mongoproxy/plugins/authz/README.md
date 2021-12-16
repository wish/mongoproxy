# authz

Authz is an authorization plugin that maps mongo commands to a Global/DB/Collection/Field CRUD operation and enforces policies using authzlib.


Notes:
- doing an un-projected read on a collection requires collection level perms or `*` field within the collection


Authorized Commands:
| Command         	| CRUD               	| Level               	|
|-----------------	|--------------------	|---------------------	|
| aggregate       	| Read               	| DB/Collection       	|
| collstats       	| Read               	| Collection          	|
| count           	| Read               	| Collection          	|
| create          	| Create             	| DB                  	|
| createIndexes   	| Create             	| DB                  	|
| currentOp       	| Read               	| Global              	|
| delete          	| Delete             	| Collection          	|
| deleteIndexes   	| Delete             	| Collection          	|
| distinct        	| Read               	| Field               	|
| dropDatabase    	| Delete             	| Global              	|
| drop            	| Delete             	| DB                  	|
| dropIndexes     	| Delete             	| Collection          	|
| endSesions      	| Delete             	| Global              	|
| explain         	| Read               	| Collection          	|
| findAndModify   	| Read/Create/Update 	| Collection/Field    	|
| find            	| Read               	| Field               	|
| getMore         	| Read               	| (same as initial Q) 	|
| hostInfo         	| Read               	| Global              	|
| insert          	| Create             	| Collection          	|
| killAllSessions 	| Delete             	| Global              	|
| killCursors     	| Delete             	| Global              	|
| killop          	| Delete             	| Global              	|
| listCollections 	| Read               	| DB                  	|
| listDatabases   	| Read               	| Global              	|
| listIndexes     	| Read               	| Collection          	|
| serverStatus    	| Read               	| Global              	|
| shardCollection   | Update               	| Global              	|
| update          	| Create/Update      	| Collection/Field    	|

OPEN_COMMAND / Unauthorized Commands:
- connectionStatus
- saslStart
- getnonce
- logout
- ping
- isMaster
- ismaster
- buildInfo
- buildinfo

TODO:
- mapReduce (block)
- validate (block)
