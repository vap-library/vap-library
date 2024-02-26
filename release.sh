#!/bin/bash

rm -rf release
mkdir release

# concatenates the source file to the destination file
# also inserts the '---' separator
function concat() {
  if [ -e $1 ]; then
    cat $1 >> $2
    echo "---" >> $2
  fi
}

# for each folder in policies
for d in policies/*; do
  # if it is a directory
  if [ -d "$d" ]; then
    concat $d/policy.yaml release/policies.yaml
    concat $d/binding.yaml release/bindings.yaml
    concat $d/crd-parameter.yaml release/crds.yaml
  fi
done

