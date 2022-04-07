#!/bin/bash


if [ "$1" == "with_build" ]; then
	rm -f app && ln -s p4click_apps/acl ./app
	make _mvn_package

	rm -f app && ln -s p4click_apps/l2forwarding ./app
	make _mvn_package

	rm -f app && ln -s p4click_apps/l3forwarding ./app
	make _mvn_package
fi

#sshpass -p "p4" scp /home/sdn/ngsdn-tutorial/p4click/acl/target/p4click-aclhandler-1.0-SNAPSHOT.oar p4@10.1.1.4:/home/p4/ngsdn-tutorial/myapps

#sshpass -p "p4" scp /home/sdn/ngsdn-tutorial/p4click/l2forwarding/target/p4click-l2forwarding-1.0-SNAPSHOT.oar p4@10.1.1.4:/home/p4/ngsdn-tutorial/myapps

#sshpass -p "p4" scp /home/sdn/ngsdn-tutorial/p4click/l3forwarding/target/p4click-l3forwarding-1.0-SNAPSHOT.oar p4@10.1.1.4:/home/p4/ngsdn-tutorial/myapps
