<?php
for ($x = 0; $x <= 1000; $x++) {
  sleep(1);
  fwrite(STDOUT, "stdout write ");
  error_log("The number is: $x", 0);
}
?>
