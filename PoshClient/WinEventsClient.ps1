function get-EventsToHash{
    param(
        [Parameter(Mandatory,Position=0)]
        [String]$CategoryName,
        [Parameter(Mandatory,Position=1)]
        [String]$Level
    )
    <#
    Level 1 - Critical
    Level 2 - Error
    Level 3 - Warning
    #################
    CategoryName - 'System' or 'application'
    #>    
    switch($level){
        'critical' {
            $EvLD = Get-WinEvent -FilterHashTable @{LogName=$categoryName; Level=1; StartTime=((get-date).AddDays(-1))} -ErrorAction SilentlyContinue
            }
        'error' {
            $EvLD = Get-WinEvent -FilterHashTable @{LogName=$categoryName; Level=2; StartTime=((get-date).AddDays(-1))} -ErrorAction SilentlyContinue
            }
        'warning' {
            $EvLD = Get-WinEvent -FilterHashTable @{LogName=$categoryName; Level=3; StartTime=((get-date).AddDays(-1))} -ErrorAction SilentlyContinue
            }
        }    

    
    #For null category events and list organize
    $zeroEvent = @{}
    $zeroEvent['count'] = 0
    $zeroEvent['id'] = ' '
    $zeroEvent['source'] = ' '
    $zeroEvent['user'] = ' '
    $zeroEvent['description'] = ' '
    # $zeroEvent['machineName'] = ' '
    $zeroEvent['dateNtime'] = ' '


    if($EvLD -eq $null){
        
        $eventsList = @()
        $eventsList += $zeroEvent
        #for [] in JSON
        $eventsList += $zeroEvent

        return($eventsList)
    }
    #counts by id
    $grouped = $EvLD | Group-Object -Property id -NoElement  
    $eventsList = @()
    foreach($eventId in $grouped.name){
        $eventHash = @{}
        
        $event = $EvLD | ? {$_.id -eq $eventId} | select -First 1

        $eventHash['count'] = ($grouped | ? {$_.name -eq $eventId}).Count
        $eventHash['id'] = $eventId
        $eventHash['Source'] = $event.ProviderName
        try{
            $e_user_sid = new-object System.Security.Principal.SecurityIdentifier(($event.userid).Value)
            $eventHash['user'] = $e_user_sid.translate([System.Security.Principal.NTAccount]).value
        }
        catch{
            $eventHash['user'] = "NA"
        }
        $eventHash['description'] = $event.message
        # $eventHash['machineName'] = $event.machineName
        $eventHash["dateNtime"] = [string]($event.timecreated.year)+":"+`
                                    [string]($event.timecreated.month)+":"+`
                                    [string]($event.timecreated.day)+"-"+`
                                    [string]($event.timecreated.hour)+":"+`
                                    [string]($event.timecreated.minute)+":"+`
                                    [string]($event.timecreated.second)
        $eventsList += $eventHash
    }
    #for [] in JSON
    $eventsList += $zeroEvent
    return($eventsList)
}

############################START SCRIPT


$computername = ((gwmi win32_computersystem -ErrorAction SilentlyContinue).name).toupper()
$ip = (gwmi win32_networkadapterconfiguration).ipaddress | ? {$_ -like "10.*"}
if(($computername -eq $null) -or ($ip -eq $null)){
    $computername = ($env:COMPUTERNAME).toupper()
    $ip = '127.0.0.1'
}

########################### Out File
$date = Get-Date
$DestStorage = "C:\junk"
$fileNameDate = [string]$date.Year+'_'+[string]$date.month+'_'+[string]$date.Day
$outFile = $DestStorage+"\allwsmonitor_incoming\"+$computername+"\"+$fileNameDate+".json"
if(!(test-path ($DestStorage + "\allwsmonitor_incoming\"+$computername+"\"))){
    try{
        New-Item -ItemType Directory ($DestStorage + "\allwsmonitor_incoming\"+$computername+"\") -InformationAction SilentlyContinue
    }catch{
        return "writeNo"
        exit
    }
}


####### Prepare Final Hash

$toJson = @{}
$toJson['computer'] = $computername
$toJson['ip'] = $ip
##
$toJson['System_Critical'] = get-eventstohash -CategoryName 'system' -Level 'critical'
$toJson['System_Error'] = get-EventsToHash -CategoryName 'system' -Level 'error'
$toJson['System_Warning'] = get-EventsToHash -CategoryName 'system' -Level 'warning'
##
$toJson['Applications_Critical'] = get-eventstohash -CategoryName 'application' -Level 'critical'
$toJson['Applications_Error'] = get-EventsToHash -CategoryName 'application' -Level 'error'
$toJson['Applications_Warning'] = get-EventsToHash -CategoryName 'application' -Level 'warning'

####### Convert To JSON and return code to SCCM BaseLine
try{
    $toJSON | ConvertTo-Json | Out-File $outFile -Encoding utf8 -Force
    return "writeYes"
}catch{
    return 'writeNo'
}