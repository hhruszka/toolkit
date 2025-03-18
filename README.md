# bptvnftester
**ATTENTION!!!**
Latest version of **btpvnftester** can be downloaded from Espoo 2 Artifactory using the following command:
```
curl https://artifactory-espoo2.int.net.nokia.com:443/artifactory/cns-bpt-tools-local/bptvnftester/latest/bptvnftester-linux-amd64 -o bptvnftester
```
**bptvnftester** is a dedicated baseline penetration testing tool for VNFs, linux based bare metals and k8s nodes. It covers 58 test cases. 
It is recommended to run it from the root account since it will automatically select all user accounts (uid >= 1000) with shell enabled, and 
it will execute test cases in the context of all selected user accounts. When it is run on a non-root account then it will test the system in the 
context of this particular account. It also allows to execute selected test cases e.g. as part of a retest effort.

It reports test result in the text format or as a json object. Additionally, it can limit reporting to failed test cases only and also provide 
additional execution details that might be used to open fault tickets for products.

Complete description of test cases executed by bptcnftester can be found in "**BPT VNF Test Plan.docx**" located the following sharepoint library: [BPT Test Plans](https://nokia.sharepoint.com/sites/cns-cn-tp-penetration-testing/BPT%20Test%20Plans/Forms/AllItems.aspx)


### Usage:
```
Usage:
  bptvnftester-linux-amd64 [flags] [test ids]

Flags:
      --detailed-report   create a report with execution details for ticketing
      --failed-only       create a report with failed only test cases
  -h, --help              help for bptvnftester-linux-amd64
  -l, --list              list all test cases
  -o, --output string     report format: text, or json (default "text")
  -t, --timeout int       timeout in seconds (default 300)
  -v, --version           prints bptvnftester-linux-amd64 version
```

### Examples:

List test available cases
```
bptvnftester -l
bptvnftester --list
```

Execute all test cases
```
bptvnftester
```

Execute selected test cases
```
bptvnftester VNFBPT31,VNFBPT41
bptvnftester VNFBPT31 VNFBPT41 VNFBPT44
```

Execute all test cases and get json formated report.
```
bptvnftester -o json
```

Execute all test cases and view only failed ones.
```
bptvnftester --failed-only
```

Execute all test cases and view only failed ones with the execution details.
```
bptvnftester --failed-only --detailed-report
```

Execute selected test cases and view only failed ones with the execution details.
```
bptvnftester --failed-only --detailed-report VNFBPT31 VNFBPT41 VNFBPT44
bptvnftester --failed-only --detailed-report VNFBPT31,VNFBPT41,VNFBPT44
```
