package rcon

import mcrcon "github.com/jlmeeker/mc-rcon"

// Send sends a message to the server's rcon
func Send(msg, port, pass string) (string, error) {
	conn := new(mcrcon.MCConn)
	err := conn.Open("localhost:"+port, pass)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	err = conn.Authenticate()
	if err != nil {
		return "", err
	}

	resp, err := conn.SendCommand(msg)
	if err != nil {
		return "", err
	}
	return resp, nil
}

/*
advancement (grant|revoke)
attribute <target> <attribute> (base|get|modifier)
ban <targets> [<reason>]
ban-ip <target> [<reason>]
banlist [ips|players]
bossbar (add|get|list|remove|set)
clear [<targets>]
clone <begin> <end> <destination> [filtered|masked|replace]
data (get|merge|modify|remove)
datapack (disable|enable|list)
debug (report|start|stop)
defaultgamemode (adventure|creative|spectator|survival)
deop <targets>
difficulty [easy|hard|normal|peaceful]
effect (clear|give)
enchant <targets> <enchantment> [<level>]
execute (align|anchored|as|at|facing|if|in|positioned|rotated|run|store|unless)
experience (add|query|set)
fill <from> <to> <block> [destroy|hollow|keep|outline|replace]
forceload (add|query|remove)
function <name>
gamemode (adventure|creative|spectator|survival)
gamerule (announceAdvancements|commandBlockOutput|disableElytraMovementCheck|disableRaids|doDaylightCycle|doEntityDrops|doFireTick|doImmediateRespawn|doInsomnia|doLimitedCrafting|doMobLoot|doMobSpawning|doPatrolSpawning|doTileDrops|doTraderSpawning|doWeatherCycle|drowningDamage|fallDamage|fireDamage|forgiveDeadPlayers|keepInventory|logAdminCommands|maxCommandChainLength|maxEntityCramming|mobGriefing|naturalRegeneration|randomTickSpeed|reducedDebugInfo|sendCommandFeedback|showDeathMessages|spawnRadius|spectatorsGenerateChunks|universalAnger)
give <targets> <item> [<count>]
help [<command>]
kick <targets> [<reason>]
kill [<targets>]
list [uuids]
locate (bastion_remnant|buried_treasure|desert_pyramid|endcity|fortress|igloo|jungle_pyramid|mansion|mineshaft|monument|nether_fossil|ocean_ruin|pillager_outpost|ruined_portal|shipwreck|stronghold|swamp_hut|village)
locatebiome <biome>
loot (give|insert|replace|spawn)
me <action>
msg <targets> <message>
op <targets>
pardon <targets>
pardon-ip <target>
particle <name> [<pos>]
playsound <sound> (ambient|block|hostile|master|music|neutral|player|record|voice|weather)
recipe (give|take)
reload
replaceitem (block|entity)
save-all [flush]
save-off
save-on
say <message>
schedule (clear|function)
scoreboard (objectives|players)
seed
setblock <pos> <block> [destroy|keep|replace]
setidletimeout <minutes>
setworldspawn [<pos>]
spawnpoint [<targets>]
spectate [<target>]
spreadplayers <center> <spreadDistance> <maxRange> (under|<respectTeams>)
stop
stopsound <targets> [*|ambient|block|hostile|master|music|neutral|player|record|voice|weather]
summon <entity> [<pos>]
tag <targets> (add|list|remove)
team (add|empty|join|leave|list|modify|remove)
teammsg <message>
teleport (<destination>|<location>|<targets>)
tell -> msg
tellraw <targets> <message>
time (add|query|set)
title <targets> (actionbar|clear|reset|subtitle|times|title)
tm -> teammsg
tp -> teleport
trigger <objective> [add|set]
w -> msg
weather (clear|rain|thunder)
whitelist (add|list|off|on|reload|remove)
worldborder (add|center|damage|get|set|warning)
xp -> experience
*/
