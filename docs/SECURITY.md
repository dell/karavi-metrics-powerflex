# Security Policy

This repository is inspected for security vulnerabilities via [gosec](https://github.com/securego/gosec) in the ```check``` target in the [Makefile](../Makefile).

Every issue detected by `gosec` is mapped to a [CWE (Common Weakness Enumeration)](http://cwe.mitre.org/data/index.html) which describes in more generic terms the vulnerability. The exact mapping can be found [here](https://github.com/securego/gosec/blob/master/issue.go#L49). The list of rules checked by `gosec` can be found [here](https://github.com/securego/gosec#available-rules).

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
|         |

## Reporting a Vulnerability

Please report a vulnerability by opening an [issue](https://github.com/dell/karavi-powerflex-metrics/issues) in this repository.
