import * as React from "react";

import { MatchCard } from "./match-card";
import { Card, Menu, Grid, Header, Image, Dropdown } from "semantic-ui-react"

export class MatchList extends React.Component<{}, {}> {
  render() {
    return (
      <div>
        <Grid centered>
          <Grid.Column widescreen={3} largeScreen={3} computer={4} tablet={6} mobile={8} verticalAlign="middle">
            <Image src="assets/tron-banner-full.png" />
          </Grid.Column>
          <Grid.Column widescreen={13} largeScreen={13} computer={16} tablet={16} mobile={16}>
            <Grid centered>
              <Grid.Column widescreen={4} computer={4} tablet={8} mobile={10}>
                <MatchCard />
              </Grid.Column>
              <Grid.Column widescreen={4} computer={4} tablet={8} mobile={10}>
                <MatchCard />
              </Grid.Column>
              <Grid.Column widescreen={4} computer={4} tablet={8} mobile={10}>
                <MatchCard />
              </Grid.Column>
              <Grid.Column widescreen={4} computer={4} tablet={8} mobile={10}>
                <MatchCard />
              </Grid.Column>
            </Grid>
          </Grid.Column>
        </Grid>
      </div>
    );
  }
}