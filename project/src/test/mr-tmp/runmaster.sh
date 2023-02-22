(cd ../../mrapps/wc && go build $RACE -buildmode=plugin wc.go) || exit 1
(cd ../../apps/mrmaster && go build $RACE mrmaster.go) || exit 1
../../apps/mrmaster/mrmaster ../data/pg*txt