<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->
# Security Policy

This repository is inspected for security vulnerabilities via [gosec](https://github.com/securego/gosec) in the ```check``` target in the [Makefile](../Makefile).

Every issue detected by `gosec` is mapped to a [CWE (Common Weakness Enumeration)](http://cwe.mitre.org/data/index.html) which describes in more generic terms the vulnerability. The exact mapping can be found [here](https://github.com/securego/gosec/blob/master/issue.go#L49). The list of rules checked by `gosec` can be found [here](https://github.com/securego/gosec#available-rules).

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
|         |

## Reporting a Vulnerability

Have you discovered a security vulnerability in Karavi PowerFlex Metrics? 
We ask you to alert the maintainers by sending an email, describing the issue, impact, and fix - if applicable.

You can reach the Karavi maintainers at karavi@dell.com.
