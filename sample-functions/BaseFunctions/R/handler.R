#!/usr/bin/env Rscript

argv <- commandArgs(TRUE)

get_stdin <- function(){
   cat('Hi', argv[1], "\n")
}

get_stdin()