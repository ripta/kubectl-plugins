#!/usr/bin/env python3
#
# Given a Godeps.json to STDIN, this script will attempt to normalize package
# names by requesting the ?go-get=1 URL for the repo and parsing its go-imports
# meta tag, or infer it from the URL.
#
# The script will then build a set of [[overrides]] sections for inclusion in 
# a Gopkg.toml file.

import html.parser
import json
import os 
import sys
import urllib.request

class GoMetaParser(html.parser.HTMLParser):
    def __init__(self):
        self.meta_values = None
        super(GoMetaParser, self).__init__()

    def handle_starttag(self, tag, attrs):
        if tag != 'meta':
            return
        attrs = dict(attrs)
        if 'name' not in attrs:
            return
        if 'go-import' != attrs['name']:
            return 
        self.meta_values = attrs['content'].replace("\\n", " ").split(' ')

def die(msg):
    info(msg)
    sys.exit(1)

def info(msg):
    print(msg, file=sys.stderr)

def normalize(repo):
    try:
        if repo.startswith("https://github.com/"):
            clean = repo.replace("https://", "").replace("?go-get=1", "")
            return "/".join(clean.split("/")[:3])
        opener = urllib.request.build_opener(urllib.request.HTTPRedirectHandler)
        contents = opener.open(repo).read()
        p = GoMetaParser()
        p.feed(str(contents))
        if p.meta_values is None:
            return None
        return p.meta_values[0].strip()
    except Exception as e:
        info("Failed when requesting " + repo)
        raise e

data = json.load(sys.stdin)
if 'Deps' not in data:
    die("Invalid data: expected 'Deps' key in data")

overrides = {}
for dep in data['Deps']:
    if 'ImportPath' not in dep:
        die("Invalid data: expected 'ImportPath' key in dep")
    if 'Rev' not in dep:
        die("Invalid data: expected 'Rev' key in dep")
    found = False
    for crepo, crev in overrides.items():
        if dep['ImportPath'].startswith(crepo + "/"):
            info(dep['ImportPath'] + " -> " + crepo + " (inferred)")
            found = True
    if found:
        continue

    repo = normalize("https://" + dep['ImportPath'] + "?go-get=1")
    if repo is None:
        info(dep['ImportPath'] + " -> ???")
        repo = dep['ImportPath']
    else:
        info(dep['ImportPath'] + " -> " + repo)
    if repo in overrides:
        if overrides[repo] != dep['Rev']:
            die("Mismatch: repository " + repo + " already pointed to revision " + overrides[repo] + " but now needs " + dep['Rev'])
    overrides[repo] = dep['Rev']

for repo, rev in overrides.items():
    print("[[override]]")
    print('  name = "{}"'.format(repo))
    print('  revision = "{}"'.format(rev))