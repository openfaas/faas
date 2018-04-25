#!/bin/sh

# Check if a user is root or is in the docker group
# The exit code is 0 if the user has permissions
# to run docker commands
# The exit code 1 if the user has to run docker
# commands using sudo

retval=1
if [ $(id -u) == 0 ]; then
  retval=0
else
  for group in $(groups)
  do
    if [ $group == 'docker' ]; then
      retval=0
      break
    fi
  done
fi
exit $retval
