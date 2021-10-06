<?php
  sleep(1);
  $d = getenv('foo');
  error_log("The number is: $d", 0);
?>
