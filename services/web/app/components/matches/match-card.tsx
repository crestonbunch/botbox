import * as React from "react";

import { Card, Icon, Button, Label } from "semantic-ui-react"
import { Versus } from "./versus";

export class MatchCard extends React.Component<{}, {}> {
  render() {
    return (
      <Card>
        <Card.Content>
          <Versus />
        </Card.Content>
        <Button attached="bottom">
          <Icon name="eye" />
          Watch
        </Button>
      </Card>
    );
  }
}

