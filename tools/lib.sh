#!/bin/bash

# elementIn checks whether a element is in an array.
#
# Args:
#   $1: The element
#   $2: The array
#
# Returns:
#   0 if the element is in the array; 1 otherwise.
function elementIn() {
  local e match="$1"
  shift
  for e; do [[ "$e" == "$match" ]] && return 0; done
  return 1
}
