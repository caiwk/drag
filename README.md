cp aaa /dev/shm 

start three node by typing `./cfs -nodeid xxx`

you will see syncpropose time out in db raft groups

if delete the code at 175 line of fop/storage.go , every thing gose well  
