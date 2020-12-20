#!/usr/bin/bash

set -eux pipefail

echo "1. check if service is running"
curl localhost:8000/health

echo "2. list all currently saved images"
curl localhost:8000/images | jq '.ImageIDList[]'

echo "3. add new images and grab them"
for png in test/data/*.png; do
  echo "sending request for image $png"
  response=$(curl -XPOST --data-binary @${png} localhost:8000/images)
  id=$(echo $response | jq -r '.ImageID')
  echo "grabbing ascii image for $id"
  imageresponse=$(curl localhost:8000/images/"${id}")
  image=$(echo $imageresponse | jq -r ".ASCIIValue")
  echo $image
done

echo "4. add new images using async and grab them"
for png in test/data/*.png; do
  echo "sending request for image $png"
  response=$(curl -XPOST -H "async: true" --data-binary @${png} localhost:8000/images)
  id=$(echo $response | jq -r '.ImageID')
  echo "grabbing ascii image for $id"
  while true; do
      imageresponse=$(curl localhost:8000/images/"${id}")
      finished=$(echo $imageresponse | jq ".Finished")
      if [[ "${finished}" == "true" ]]; then
        image=$(echo $imageresponse | jq -r ".ASCIIValue")
        echo $image
        break
      fi
      echo "image not yet generated...wait for a bit"
      sleep 5
  done
done

echo "5. send bad data"
for png in test/baddata/*; do
  echo "sending request for image $png"
  curl -XPOST --data-binary @${png} localhost:8000/images
done

echo "6. list all currently saved images again"
curl localhost:8000/images | jq '.ImageIDList[]'

echo "done"