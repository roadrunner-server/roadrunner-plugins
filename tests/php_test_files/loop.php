<?php
for ($x = 0; $x <= 5; $x++) {
  sleep(1);
  error_log("The number is: $x", 0);
}
?>
