question: """PUT /
            Type your name:
            """ @ stdout;
answer: "GET /" @ stdin;

file: "GET /" @ file:///$HOME/file.txt

read: {"GET /${it}" @ file:///$PWD/}