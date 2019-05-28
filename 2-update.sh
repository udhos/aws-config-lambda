#!/bin/sh

aws lambda update-function-code --function-name FunctionConfigLambda --zip-file fileb://main.zip
