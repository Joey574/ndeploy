# Core Design Concepts

### What does this tool do?
NDpeloy is a nixos gui fleet management tool over ssh for homelab and large deployments of nixos

### General Flow
* On startup verify permissions (need to be trusted nixos user) and prompt for login
* Great with main menu, maybe on first start boot straight to a tutorial or add first server page?
* Define servers by user, host, ssh key path
* Test connection to server and verify permissions
* Verify remote is nixos
* Build database (sqlite?) of remote servers, need to enumerate information like cpu, rma, gpu, ssd, etc
* Deploy systemd unit for remote to gather metrics and other information
* Allow user to mark as a build server, will need to determine what to do with multiple build servers
* -- Q: Should build servers automatically expose a harmonia endpoint so we don't waste cache when another server has the information? should this be an on by default option? Do we make a gcroot by default for the remote server?

### Goals
* Provide a intuitive gui for deployming and monitoring the status of remote servers
* Establish remote systemd unit for gathering systeminfo
* Grouping of servers
* Staged deployments for grouped servers
* Fine tuned control for the deployment process and ability to hook into custom tools
* Automatic rollback and alert integration for failed servers
* Caching control for build servers - maybe build fleets? ability to distribute compilation jobs across existing nodes
* Single user / operator
* Provide node collector

### Approach
* VM Testing environment for local development and PoC
