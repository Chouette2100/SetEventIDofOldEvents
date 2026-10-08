#!/bin/bash
filename=`date +%Y%m%d-%H%M`
tar zcvf SetEventIDofOldEvents_$filename.tar.gz DBConfig.yml SetEventIDofOldEvents run.sh
