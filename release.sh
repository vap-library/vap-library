#!/bin/bash

for onefile in policies bindings crds;do
  rm -f release/${onefile}.yaml
done

function remove_leading_separator() {
    if [[ $(head -n 1 "$1") == "---" ]]; then
        tail -n +2 "$1"
    else
        cat "$1"
    fi
}

# concatenates the source file to the destination file
# also inserts the '---' separator
function concat() {
  if [ -e $1 ]; then
    content=$(remove_leading_separator $1)
    echo "$content" >> $2
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

