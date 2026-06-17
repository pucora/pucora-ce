---
name: Help and questions
about: You are stuck trying to do something, you get unexpected results, or you have a question or suggestion
title: ''
labels: 'question'
assignees: ''

---
<!--
Thank you for using Velonetics! Please spend some time to fill all the requested information in this template.

Having the proper context and detailed information will help us provide a faster answer. Unfortunately, we have to leave issues that we don't wholly understand or require more information for a much later processing.
-->

**Environment info:**

* Velonetics version: Run `velonetics version` and copy the output here
* System info: Run `uname -srm` or write `docker` when using containers
* Hardware specs: Number of CPUs, RAM, etc
* Backend technology: Node, PHP, Java, Go, etc.
* Additional environment information:

**Describe what are you trying to do**:
A clear and concise description of what you want to do and what is the expected result.

**Your configuration file**:
<!-- The content of your `velonetics.json`. When using the flexible configuration option, the computed file can be generated specifying the env var FC_OUT=out.json -->

```json
{
  "version": 3,
  ...
}
```

**Configuration check output**:
Result of `velonetics check -dtc velonetics.json --lint` command

```
Output of the linter here.
```

**Commands used:**
How did you start the software?
```
# Example:
velonetics run -d -c velonetics.json

# Or with Docker:
docker run --rm -it -v $PWD:/etc/velonetics \
        -e FC_ENABLE=1 \
        -e FC_SETTINGS="/etc/velonetics/config/settings" \
        -e FC_PARTIALS="/etc/velonetics/config/partials" \
        -e FC_OUT=out.json \
        niteesh20/velonetics:2.0.0 \
        run -c /etc/velonetics/config/velonetics.json -d
```

**Logs:**
Logs you saw in the console and debugging information

**Additional comments:**
