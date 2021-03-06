#!/bin/bash
#
# /etc/rc.d/init.d/esmeralda
#
# chuanyun.io esmeralda
#
#  chkconfig: 2345 20 80 Read
#  description: chuanyun.io esmeralda
#  processname: esmeralda

# Source function library.
. /etc/rc.d/init.d/functions

PROGHOME=${PWD}
PROG=${PROGHOME}/esmeralda

PROGNAME=${PROG##*/}

LOCKFILE=${PROGHOME}/$PROGNAME.lock
PIDFILE=${PROGHOME}/$PROGNAME.pid

CONFIG_FILE_PATH=${PROGHOME}/esmeralda.toml

start() {
    echo -n $"Starting $PROGNAME: "

    daemon --check=$PROGNAME "${PROG} -config=$CONFIG_FILE_PATH -pidfile=$PIDFILE -pprof=true -pprof.port=10201 &"
    RETVAL=$?
    echo
    [ $RETVAL = 0 ] && touch ${LOCKFILE}
    return $RETVAL
}

stop() {
    echo -n $"Stopping ${PROGNAME}: "
    killproc -p ${PIDFILE} ${PROGNAME}
    RETVAL=$?
    echo
    [ $RETVAL = 0 ] && rm -f ${LOCKFILE} ${PIDFILE}
}

rh_status() {
    status -p ${PIDFILE} -b ${PROG} ${PROG}
}

case "$1" in
    start)
        rh_status >/dev/null 2>&1 && exit 0
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        start
        ;;
    *)
        echo $"Usage: ${PROGNAME} {start|stop|restart}"
        exit 2
        ;;
esac