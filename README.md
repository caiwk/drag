cp aaa /dev/shm 

start three node by typing `./cfs -nodeid xxx`

hit enter on any node

you will see syncpropose time out in db raft groups

if delete the code at 175 line of fop/storage.go , every thing gose well  
