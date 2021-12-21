#!/bin/sh -ex
go-test-trace < test_report.json > test_trace.json
${CATAPULT:-$HOME/catapult}/tracing/bin/trace2html test_trace.json --output test_trace.html
