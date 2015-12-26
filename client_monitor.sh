#!/bin/bash

getCpuSample ()
{
    totalcpuusage=`top -bn 1 | awk 'NR>7{s+=$9} END {print s}'`
    cpuusage=`echo $totalcpuusage/$(nproc) | bc -l`
}

while :
do
    getCpuSample
    cpusample1=$cpuusage
    sleep 1

    getCpuSample
    cpusample2=$cpuusage
    sleep 1

    getCpuSample
    cpusample3=$cpuusage

    cpuusage=0`echo "($cpusample1+$cpusample2+$cpusample2)/300" | bc -l`




    totalmemory=`free -m | awk 'NR==2{printf "%s", $2 }'`
    usedmemory=`free -m | awk 'NR==2{printf "%s", $3 }'`

    #swapon --show=Size,Used --noheadings --bytes | while read x; do
    while read x
    do
        swaptotalmemory=$(echo $(echo "$x" | cut -d' ' -f1) "/1048576" | bc -l)
        swapusedmemory=`echo $(echo $x | cut -d' ' -f2)"/1048576" | bc -l`
        totalmemory=$(echo "$totalmemory+$swaptotalmemory" | bc -l)
        usedmemory=$(echo "$usedmemory+$swapusedmemory" | bc -l)
    done <<< "`swapon --show=Size,Used --noheadings --bytes`"

    memoryusage=0`echo "$usedmemory/$totalmemory" | bc -l`

    logger -p kern.info -t cpu_state $cpuusage
    logger -p kern.info -t memory_state $memoryusage
    sleep 1
done

