import * as React from "react";
import { browserHistory } from "react-router";
import { Segment, Container, Header, Image, Button, Divider } from "semantic-ui-react";

import { MatchList } from "../views/matches";

export class HomePage extends React.Component<{}, {}> {
  render() {
    return (
      <div>
        <div style={{ backgroundImage: "url('assets/bg-tiles-blue.png')" }}>
          <Segment vertical basic textAlign="center">
            <Container text>
              <Image size="medium" src="assets/botbox-masthead.svg" centered />
              <Header as="h1">Join the sport of programming AI</Header>

              <Divider hidden />
              <Button primary size="massive" onClick={() => browserHistory.push('/register')}>Start playing</Button>
              <Divider hidden />
            </Container>
          </Segment>
        </div>
        <Segment vertical basic padded="very">
          <Container>
            <MatchList />
          </Container>
        </Segment>
      </div>
    );
  }
}
