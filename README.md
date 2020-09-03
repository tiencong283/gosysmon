# Gosysmon

Gosysmon is an security endpoint detection platform (a module of [EDR](https://en.wikipedia.org/wiki/Endpoint_Detection_and_Response) system) where it uses [Sysmon](https://docs.microsoft.com/en-us/sysinternals/downloads/sysmon) as its sensor.

It mainly focuses on adversary behaviors of advanced attacks so it implemented the [MITRE ATT&CK framework](https://attack.mitre.org/) and use it as the language for describing suspicious and malicious behaviors.

## Primary features

It currently offers:
- Multi-layered detection including: IOC filter for known attacks, rule-based and model-based behavior filter for unknown attacks
- Deep visibility of alerts for inspection and analysis of threats
- Ability to detect similar processes which shares the same suspicious behaviors as the known ones 

It have been covered [MITRE ATT&CK techniques](https://mitre-attack.github.io/attack-navigator/enterprise/#layerURL=https://raw.githubusercontent.com/tiencong283/gosysmon/master/rules/layer.json?token=AGKK7WQXIIP6DBJ6KLSK27263BXN6)

Note: **It supports only Windows client**

## Installation
For server side see [server installation guide](https://github.com/tiencong283/gosysmon/wiki/Server-Installation-Guide)

For client side see [client installation guide](https://github.com/tiencong283/gosysmon/wiki/Client-Installation-Guide)
## Related projects
* [MITRE ATT&CK](https://attack.mitre.org/)
* [MITRE ATT&CK Evaluations](https://attackevals.mitre.org/)
* [Atomic Red Team](https://github.com/redcanaryco/atomic-red-team)
* [SwiftOnSecurity/sysmon-config](https://github.com/SwiftOnSecurity/sysmon-config)

## Feedback
I'd appreciate any kind of feedback on this project, please [opening an issue](https://github.com/tiencong283/gosysmon/issues/new) or [contact me directly](mailto:tiencong283@outlook.com?subject=[Feedback%20About%20Gosysmon])