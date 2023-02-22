# MIT 6.824 2020

the Lec record and Lab implement for [6.824 Schedule: Spring 2020](http://nil.csail.mit.edu/6.824/2020/schedule.html)

change the project implement to go 1.17 (the original project on the website is for go 1.13).

## Lab 1  MapReduce

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


