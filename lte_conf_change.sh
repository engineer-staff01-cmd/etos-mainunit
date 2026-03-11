#!/usr/bin/bash
sed -i -e "/^FORCE_REBOOT=/c\FORCE_REBOOT=TRUE" /etc/connection-recover/gsm-ttyACM0_connection-recover.conf
#sed -i -e "/^FORCE_REBOOT=/c\FORCE_REBOOT=FALSE" /etc/connection-recover/gsm-ttyACM0_connection-recover.conf   # default
sed -i -e "/^FORCE_RECONNECT_PING_NG_COUNT=/c\FORCE_RECONNECT_PING_NG_COUNT=5" /usr/bin/connection-recoverd
#sed -i -e "/^FORCE_RECONNECT_PING_NG_COUNT=/c\FORCE_RECONNECT_PING_NG_COUNT=2" /usr/bin/connection-recoverd   # default
sed -i -e "/^FORCE_REBOOT_PING_NG_COUNT=/c\FORCE_REBOOT_PING_NG_COUNT=\$\(\(\ \$FORCE_RECONNECT_PING_NG_COUNT+3\ \)\)" /usr/bin/connection-recoverd
#sed -i -e "/^FORCE_REBOOT_PING_NG_COUNT=/c\FORCE_REBOOT_PING_NG_COUNT=\$\(\(\ \$FORCE_RECONNECT_PING_NG_COUNT+2\ \)\)" /usr/bin/connection-recoverd   # default
systemctl stop connection-recover.service
systemctl start connection-recover.service
