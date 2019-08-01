#
# Copyright (C) 2019 Nalej Group - All Rights Reserved
#

# iron/go is the alpine image with only ca-certificates added
FROM iron/go

RUN mkdir /nalej

# Now just add the binary
COPY metrics-collector /nalej/

ENTRYPOINT ["./nalej/metrics-collector"]
