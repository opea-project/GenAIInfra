#!/bin/bash

sudo apt-get install -y dockder.io git
sudo snap install helm --classic
wget https://arxiv.org/pdf/2212.04088 -o 2212.pdf
rm -rf GenAIInfra
git clone https://github.com/opea-project/GenAIInfra/
cd GenAIInfra; git checkout 66de41c00950898d13811d6df7383bdac50c26ca
