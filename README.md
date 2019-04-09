# lightning-poll
Lightning poll provides users with the ability to create refundable polls using hodl invoices. It was my entry for the 2019 Boltathon hackathon.


Poll creators can choose a refund strategy for polls they create, refunding the majority of voters, minority or voters, all or none. When a poll closed, users matching the refund strategy are refunded, and the poll creator is paid out the remaining total.


# Install
A [LND node](https://github.com/lightningnetwork/lnd/blob/master/docs/INSTALL.md) and [golang](https://golang.org/doc/install) installation are required to run lightning-poll. 

`go install $GOPATH/lightning-poll`
`$GOPATH/bin/lightning-poll --lnd_cert={lnd cert path} --lnd_address{lnd rpc server}` 
