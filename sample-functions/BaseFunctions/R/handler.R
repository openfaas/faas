#!/usr/bin/env Rscript

f <- file("stdin")
open(f)
line<-readLines(f, n=1, warn = FALSE)

write(paste0("Hi ", line), stderr())
