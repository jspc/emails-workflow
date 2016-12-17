emails-workflow
==

Simple workflow to collate useful information for the day ahead and email it to me.

![screenshot of an email sent from this workflow](doc/sample-email.png)


| who       | what |
|-----------|------|
| dockerhub | https://hub.docker.com/r/jspc/emails-workflow/   |
| circleci  | https://circleci.com/gh/jspc/emails-workflow   |
| licence   | MIT   |


Running
--

This package provides a customised `JobManager`, as set out in https://github.com/gincorp/sample-job-runner

It relies on some env vars:

```bash
$ export OWM_API_KEY='djdjdjdjdj'        # A relic from https://github.com/briandowns/openweathermap
$ export MAIL_USERNAME='foo'             # username with which to connect to mail server
$ export MAIL_PASSWORD='bar'             # password for the same
```

It is run like any gin command:

```
$ ./emails-workflow -h
Usage of ./emails-workflow:
  -amqp string
        URI to pass messages via (default "amqp://guest:guest@localhost:5671/")
```

Configuration
--

Workflow creation lives in `script/wf.go` and relies on two env vars being set:

```bash
$ export SENDER_ADDRESS=me@example.com            # From address on email
$ export RECEIPIENT_ADDRESS=someone@example.com   # To address on email
$ export GUARDIAN_API_KEY=jsjsjsjsjs              # A valid api key for the guardian API
```

It *can* be compiled and run:

```bash
$ cd script/
$ go build
$ ./script
```

But this leads to an absurdity in the name and will muck up if more scripts are added. Thus:

```bash
$ go run script/wf.go
```

Licence
--

MIT License

Copyright (c) 2016 jspc

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
