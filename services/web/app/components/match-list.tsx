import * as React from "react";

import { MatchCard } from "./match-card";
import { Card, Menu, Divider, Header, Image, Dropdown } from "semantic-ui-react"

export class MatchList extends React.Component<{}, {}> {
  render() {
    return (
      <div>
      <Menu secondary>
        <Menu.Item header>
          <Image size="small" src="assets/Tron Banner.svg" />
        </Menu.Item>

        <Dropdown as={Menu.Item} text="Recent matches">
          <Dropdown.Menu>
            <Dropdown.Item text="Popular matches" />
          </Dropdown.Menu>
        </Dropdown>

        <Menu.Menu position="right">
          <Menu.Item>See more</Menu.Item>
        </Menu.Menu>

      </Menu>
      <Divider hidden />
      <Card.Group stackable itemsPerRow="four">
        <MatchCard />
        <MatchCard />
        <MatchCard />
        <MatchCard />
      </Card.Group>
      </div>
    );
  }
}

