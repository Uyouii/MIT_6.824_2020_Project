# MIT 6.824 2020

This Project includes the Lec record and Lab implementation for [6.824 Schedule: Spring 2020](http://nil.csail.mit.edu/6.824/2020/schedule.html) .

The MIT 6.824 is a very amazing course, even after working for five years, a lot of content is still worth learning.

I change the project implementation to go 1.17 (the original project on the website is for go 1.13).

The directory of this repositories:

- [project](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/project): project is for the lab and code
- [tutorial](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/): tutorial record the leasons file and notes.

The Reading Records are in [BookReading Project](https://github.com/Uyouii/BookReading/tree/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F)


## Lecs

| Name                                                         | Record                                                       | Papers                                                       | Reading Recored                                              |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| [LEC1 introduction](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC1%20introduction) | [notes.txt](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC1%20introduction/notes.txt) | [MapReduce](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/MapReduce/mapreduce.pdf) | [MapReduce: Simplified Data Processing on Large Clusters(2004)](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/MapReduce/MapReduce:%20Simplified%20Data%20Processing%20on%20Large%20Clusters.md) |
| [LEC2 RPC and Threads](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC2%20RPC%20and%20Threads) | [notes.txt](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC2%20RPC%20and%20Threads/notes.txt) |                                                              |                                                              |
| [LEC3 GFS](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC3%20GFS) | [notes](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC3%20GFS/notes.txt) | [**GFS**(The Goole File System)](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC3%20GFS/gfs%202003.pdf) | [The Google File System](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/GFS/The%20Google%20File%20System.md)<br />[GFS文件系统](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/GFS/GFS%E6%96%87%E4%BB%B6%E7%B3%BB%E7%BB%9F.md) |
| [LEC4 Primary-Backup Replication](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC4%20Primary-Backup%20Replication) | [notes](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC4%20Primary-Backup%20Replication/notes.txt) | [Fault-Tolerant Virtual Machines](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC4%20Primary-Backup%20Replication/vm-ft.pdf) | [The Design of a Practical System for Fault-Tolerant Virtual Machines(2010)](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/VM-FT/Fault-Tolerant%20Virtual%20Machines.md) |
| [LEC5 Go, Threads, and Raft](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC5%20Go%2CThreads%20and%20Raft) | [notes](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC5%20Go%2CThreads%20and%20Raft/notes.txt) |                                                              |                                                              |
| [LEC6 Fault Tolerance: Raft(1)](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC6%20Fault%20Tolerance%3A%20Raft(1)) | [notes](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC6%20Fault%20Tolerance%3A%20Raft(1)/notes.txt) | [Raft](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC6%20Fault%20Tolerance%3A%20Raft(1)/raft-extended.pdf) | [In Search of an Understandable Consensus Algorithm (Extended Version) 2014](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/Raft/In%20Search%20of%20an%20Understandable%20Consensus%20Algorithm%20(Extended%20Version).md)<br />[Raft共识算法](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/Raft/Raft%E5%85%B1%E8%AF%86%E7%AE%97%E6%B3%95.md) |
| [LEC7 Fault Tolerance: Raft (2)](http://nil.csail.mit.edu/6.824/2020/notes/l-raft2.txt), | [notes](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC7%20Fault%20Tolerance%3A%20Raft%20(2)/notes.txt) |                                                              |                                                              |

## Labs

### Lab 1  MapReduce

lab link: http://nil.csail.mit.edu/6.824/2020/labs/lab-mr.html

change the test-mr.sh and file location to adapt to go1.17

run test:

```sh
cd project/src/test
./test-mr.sh

some out:
*** Starting wc test.
--- wc test: PASS
*** Starting indexer test.
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
--- reduce parallelism test: PASS
*** Starting crash test.
--- crash test: PASS
*** PASSED ALL TESTS
```

### Lab2 : Raft

lab link: http://nil.csail.mit.edu/6.824/2020/labs/lab-raft.html

#### Part 2A

Output：

```sh
$ go test -run 2A
Test (2A): initial election ...
  ... Passed --   3.1  3   56    6950    0
Test (2A): election after network failure ...
  ... Passed --   4.4  3  120    9678    0
PASS
ok      6.824/raft      8.935s
```



