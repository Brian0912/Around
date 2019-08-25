sudo systemctl enable elasticsearch
sudo systemctl start elasticsearch

export GOOGLE_APPLICATION_CREDENTIALS="/home/bh800912/Documents/Around/Around-034ae903e269.json"
echo $GOOGLE_APPLICATION_CREDENTIALS

go run main.go user.go

