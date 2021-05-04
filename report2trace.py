#!/usr/bin/env python

# convert go test report JSON into chrome tracing JSON

import sys
import json
import datetime
import dateutil.parser


def datetime_parser(json_dict):
  for (key, value) in json_dict.items():
    try:
      json_dict[key] = dateutil.parser.parse(value)
    except (ValueError, AttributeError, TypeError):
      pass
  return json_dict


package = {}
start = {}
end = {}
result = {}
min = None

report = sys.argv[1]
with open(report) as tr:
  for line in tr:
    event = json.loads(line, object_hook=datetime_parser)
    action = event['Action']
    if action == "output":
      continue
    time = event['Time']
    if action == "run":
      test = event['Test']
      start[test] = time
      if min is None or time < min:
        min = time
      if 'Package' in event:
        pkg = event['Package']
      else:
        pkg = 'main'
      if pkg in package:
        package[pkg].append(test)
      else:
        package[pkg] = [test]
    elif action == "pass" or action == "fail" or action == "skip":
      if 'Test' not in event:
        continue
      test = event['Test']
      end[test] = time
      assert(test in start)
      result[test] = action
    elif action == "pause" or action == "cont":
      continue
    else:
      print("Unknown action: ", action)
      continue

events = []
for pkg in package:
  tests = package[pkg]
  tests.sort()
  for test in tests:
    if test not in result:
      continue
    res = result[test]
    duration = end[test] - start[test]
    # print(pkg, test, start[test], end[test], duration)
    b = int((start[test] - min).total_seconds() * 1e6)
    e = int((end[test] - min).total_seconds() * 1e6)
    events.append({"ts": b, "pid": pkg, "tid": test, "ph": "B", "name": test, "cat": res})
    events.append({"ts": e, "pid": pkg, "tid": test, "ph": "E"})
doc = {"displayTimeUnit": "ms", "traceEvents": events}
print(json.dumps(doc))
