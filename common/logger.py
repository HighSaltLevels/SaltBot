""" Logging Module """


class Logger:
    """Logger object for writing the log file"""

    def __init__(self, logfile="log.txt"):
        self._logfile = logfile
        self.initialize_logfile()

    def log(self, msg):
        """Print the message and write to the log file"""
        try:
            print(msg, flush=True)
        except UnicodeDecodeError:
            print(
                "WARN: Could not print message. User, channel or msg had unicode bytes"
            )
            print("WARN: Check the logfile", flush=True)
        with open(self._logfile, "a") as stream:
            stream.write(f"{msg}\n")

    def initialize_logfile(self):
        """Clear the log file"""
        with open(self._logfile, "w"):
            pass


LOGGER = Logger()
