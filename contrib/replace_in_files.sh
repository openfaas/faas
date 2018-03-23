#!/bin/sh

OLD_VERSION=$1
NEW_VERSION=$2

echo "Replacing templates for version $OLD_VERSION with $NEW_VERSION"

grep -r -l --fixed-strings $OLD_VERSION .*.md * | while read fname; do
    echo "Replacing version in file $fname"
    sed -i '' "s/$OLD_VERSION/$NEW_VERSION/g" $fname
done;

echo DONE Replacing version templates.
