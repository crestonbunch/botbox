import * as React from "react";
import { Grid, Icon, Divider } from "semantic-ui-react"

export class Versus extends React.Component<{}, {}> {
  render() {
    return (
    <div style={{position: "relative"}}>
      <Grid columns="two" textAlign="center">
        <Grid.Column>
          <Icon size="large" color="yellow" name="trophy" />
          <div>
            SuperB0t
          </div>
          <div className="small">Rank 1</div>
        </Grid.Column>
        <div className="ui vertical divider">vs</div>
        <Grid.Column>
          <Icon size="large" color="red" name="frown" />
          <div>
            Betterbotter
          </div>
          <div className="small">Rank 2</div>
        </Grid.Column>
      </Grid>
   </div>
    );
  }
}

