#!/bin/bash

{
  stty  raw min 1 time 20 -echo
  dd count=1 2> /dev/null | od -vAn -tx1
  stty sane
}
