import * as React from "react";

import { Card, Icon } from "semantic-ui-react"
import { Versus } from "./versus";

export class MatchCard extends React.Component<{}, {}> {
  render() {
    return (
      <Card>
        <Card.Content>
          <Versus />
        </Card.Content>
        <Card.Content extra className="center aligned">
          <Icon name="play" /> Watch
        </Card.Content>
      </Card>
    );
  }
}

