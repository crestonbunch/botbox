import * as React from "react";
import * as ReactDOM from "react-dom";
import { Button, Segment, Container, Image, Grid, Header, Icon } from "semantic-ui-react"

export class Footer extends React.Component<{}, {}> {

  render() {
      return (

        <Segment padded="very" inverted>
          <Container textAlign="center">
            <Grid>
              <Grid.Column width="ten">
                <Image size="small" src="assets/Inverted Botbox Banner.svg" />
              </Grid.Column>
              <div className="three wide column">
                <Header as="h4" inverted>Botbox is free software</Header>
                <Button size="small" inverted><Icon name="github" /> Contribute</Button>
              </div>
              <div className="three wide column">
                <Header as="h4" inverted>Servers are expensive</Header>
                <Button size="small" inverted><Icon name="heart" /> Donate</Button>
              </div>
            </Grid>
          </Container>
        </Segment>
      );
  }
}

