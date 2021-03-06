sudo /etc/init.d/loginServer stop
sudo /etc/init.d/gameServer stop

cp gk/bin/loginServerMain $HOME/loginServer/bin
cp gk/bin/gameServerMain $HOME/gameServer/bin

cp -r ../stylesheets /var/www/gourdianknot/assets/gk
cp -r ../javascript /var/www/gourdianknot/assets/gk
cp -r $HOME/assets/game/audio /var/www/gourdianknot/assets/gk
sudo cp -r ../gktool /var/www/gourdianknot

cp ../../assets/game/svg/controls/*.json /var/www/gourdianknot/assets/gk/controls

cd ../../assets/game/svg/controls
for file in *.svg
do
	/home/turnip/server/go/gk/bin/fixSvgMain -in=$file -out=/var/www/gourdianknot/assets/gk/controls/$file
done

sudo /etc/init.d/loginServer start
sudo /etc/init.d/gameServer start
