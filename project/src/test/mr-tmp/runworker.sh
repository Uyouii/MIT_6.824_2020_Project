(cd ../../mrapps/wc && go build $RACE -buildmode=plugin wc.go) || exit 1
(cd ../../apps/mrworker && go build $RACE mrworker.go) || exit 1
../../apps/mrworker/mrworker ../../mrapps/wc/wc.so