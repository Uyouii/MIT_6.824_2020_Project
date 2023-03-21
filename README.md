# MIT 6.824 2020

This Project includes the Lec record and Lab implementation for [6.824 Schedule: Spring 2020](http://nil.csail.mit.edu/6.824/2020/schedule.html) .

The MIT 6.824 is a very amazing course, even after working for five years, a lot of content is still worth learning.

I change the project implementation to go 1.17 (the original project on the website is for go 1.13).

The directory of this repositories:

- [project](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/project): project is for the lab and code
- [tutorial](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/): tutorial record the leasons file and notes.

## Papers

Below are the links to the papers included in the course.

| Name                           | Link                                                         | Reading Recored                                              |
| ------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| **MapReduce**                  | [link](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/MapReduce/mapreduce.pdf) | [reading record](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/MapReduce/MapReduce%3A%20Simplified%20Data%20Processing%20on%20Large%20Clusters.md) |
| **GFS**(The Goole File System) | [link](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC3%20GFS/gfs%202003.pdf) | [reading record](https://github.com/Uyouii/BookReading/tree/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/GFS) |

## Lecs

| Name                                                         | Record                                                       | Paper                                                        |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| [LEC1 introduction](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC1%20introduction) | [notes.txt](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC1%20introduction/notes.txt) | [mapreduce](https://github.com/Uyouii/BookReading/blob/master/%E5%88%86%E5%B8%83%E5%BC%8F%E7%B3%BB%E7%BB%9F/MapReduce/mapreduce.pdf) |
| [LEC2 RPC and Threads](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC2%20RPC%20and%20Threads) | [notes.txt](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC2%20RPC%20and%20Threads/notes.txt) |                                                              |
| [LEC3 GFS](https://github.com/Uyouii/MIT_6.824_2020_Project/tree/master/tutorial/LEC3%20GFS) | [notes](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC3%20GFS/notes.txt) | [GFS](https://github.com/Uyouii/MIT_6.824_2020_Project/blob/master/tutorial/LEC3%20GFS/gfs%202003.pdf) |

## Labs

### Lab 1  MapReduce

[link](http://nil.csail.mit.edu/6.824/2020/labs/lab-mr.html)

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

