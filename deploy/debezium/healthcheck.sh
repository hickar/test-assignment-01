#!/bin/sh

nc -z localhost 8083 && [ $(curl -s http://localhost:8083/connectors/) != '[]' ]
