import * as React from "react";
import { Segment, Container, Header, Image, Button, Divider } from "semantic-ui-react"

import { Navigation } from "./navigation";
import { MatchList } from "./match-list";
import { Footer } from "./footer";
import { Register } from "./register";

export class Home extends React.Component<{}, {}> {
  render() {
    return (
    <div>
      <div style={{backgroundImage:"url('assets/Background Tile Transparent.png')"}}>
        <Segment vertical basic textAlign="center">
          <Container>
            <Navigation />
          </Container>
          <Container text>
            <Image size="medium" src="assets/Botbox Masthead.svg" centered />
            <Header as="h1">Join the sport of programming AI</Header>

            <Divider hidden />
            <Register trigger={<Button primary size="massive">Start playing</Button>}/>
            <Divider hidden />

          </Container>
        </Segment>
      </div>
      <Segment vertical basic padded="very">
        <Container>
          <MatchList />
        </Container>
      </Segment>

      <Footer />
    </div>
    );
  }
}

