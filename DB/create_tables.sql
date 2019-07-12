-- The are 6 tables with the same columns without references or dependencies
-- may by it's ugly but I hope go func insert events in 6 tables 
-- in parallel much more easies then in one
-- and for collect machines or users info we use MS SCCM
-- WsEventsMonitor.dbo.SystemCriticals
-- WsEventsMonitor.dbo.SystemErrors
-- WsEventsMonitor.dbo.SystemWarnings
-- WsEventsMonitor.dbo.ApplicationsCriticals
-- WsEventsMonitor.dbo.ApplicationsErrors
-- WsEventsMonitor.dbo.ApplicationsWarnings

CREATE TABLE WinEventsMonitor.dbo.SystemErrors (
	id INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
	machine NVARCHAR (50) NOT NULL,
	eventid NCHAR (10) NOT NULL,
	source NVARCHAR (MAX),
	description NVARCHAR (MAX),
	count TINYINT NOT NULL,
	datentime DATETIME NULL,
	subnet2 TINYINT NOT NULL,
	subnet3 TINYINT NOT NULL,
	event_user NVARCHAR (50) NULL
)

-- And once other table is for filtering processed json-files

CREATE TABLE WinEventsMonitor.dbo.ProcessedFiles (
	id INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
	machineDir NVARCHAR (50) NOT NULL,
	fileDateName  NVARCHAR (50) NOT NULL,
	processedDate DATETIME(120) NOT NULL,
	result NCHAR (10) NOT NULL
)