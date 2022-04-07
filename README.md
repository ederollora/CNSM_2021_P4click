# CNSM_2021_P4click

Project structure published at the 17th International Conference on Network and Service Management conference.


You can reference this code or paper with this citation:
```bibtex
@INPROCEEDINGS{eoz,
  author={Ollora Zaballa, Eder and Franco, David and Jacob, Eduardo and Higuero, Marivi and Berger, Michael Stübert},
  booktitle={2021 17th International Conference on Network and Service Management (CNSM)},
  title={Automation of Modular and Programmable Control and Data Plane SDN Networks},
  year={2021},
  volume={},
  number={},
  pages={375-379},
  doi={10.23919/CNSM52442.2021.9615508}
}
```

## Code

* The code that builds the pipeline out of .p4z modules (= zip files) is in the p4click folder.
* We worked with our own repository named modules.p4.click (held module code archives in p4z format), which is hardcoded in the code. You can uncomment the LOC referring to the local "repository" folder (L74 in main).
* Most of the test code was based on the [NG-SDN tutorial](https://github.com/opennetworkinglab/ngsdn-tutorial) example. The Makefile, for instance, is a modified version of it. It hols some custom commands to build apps, deploy custom scenarios, etc.
* The "examples" folder should have some handy files like a compiled verison of the app, .oar files for ONOS, modules as p4z files, mininet topology files, ONOS network config, etc.
* Since many paths were hardcoded the code has not been tested out of the test environment. You can check the code while we test and modify some parts before there is an out-of-the-box working version that does not depend on our local test environment.
* Many features as hal-away implemented as they were not necessary for the paper but for a nice demo to show, thus, do not expect everything on the code to work: For instance, install an ONOS app from the P4click CLI. We do this from a bash script.
* P4click was configured, mainly, to work as a CLI-based app that 'built a network' out of commands, see an example here:
```
p4click > create deployment mydeployment
p4click > use deployment mydeployment
p4click/mydeployment > create switch switch1
p4click/mydeployment > use switch switch1
../mydeployment/switch1 > set model v1model
../mydeployment/switch1 > set target bmv2
../mydeployment/switch1 > add module l2forwarding
../mydeployment/switch1 > show modules
features:
- l2forwarding
../mydeployment/switch1 > add module aclhandler
../mydeployment/switch1 > add module l3forwarding
../mydeployment/switch1 > show modules
features:
- l3forwarding
- l2forwarding
- aclhandler
../mydeployment/switch1 > set pipeline-order l2forwarding,l3forwarding,aclhandler
in -> l2forwarding -> l3forwarding -> aclhandler -> output
../mydeployment/switch1 > build data plane
```
