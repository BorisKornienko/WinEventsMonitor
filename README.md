# WinEventsMonitor
## The goal of this system is a finding of anomalies in Windows Events
Powershell script get events from the machine, struct them in to the JSON file and put in to the network share.
Golang parse these files, insert data in MS SQL and remove files
From DB data can be visualizing by the Grafana, PowerBI, D3, etc.
### Visualize anomaly in medium or large environments it's a pretty important thing, as I think.
