iscsiadm -m session   |  awk '{
 if(match($2,/([0-9]+)/,a)){
  result=a[1] ",";
  }
  result=result  $4 ",";
  if(match($3,/([0-9\.]+):/,b)){
    result=result b[1] ",";
  }
  if(match($3,/:([0-9\.]+)/,b)){
      result=result b[1] ;
    }
  print result
  }'