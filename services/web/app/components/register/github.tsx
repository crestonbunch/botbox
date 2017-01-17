import * as React from "react";
import "whatwg-fetch"
import { Icon, Button } from "semantic-ui-react";
import { Settings } from "../../settings";

export interface GithubProps {
}

export interface GithubState {
}

const GITHUB_ACCESS_URL = "https://github.com/login/oauth/authorize";
const GITHUB_ACCESS_PARAMS = {
  client_id: Settings.GITHUB_ID,
  scope: "user"
}

export class Github extends React.Component<GithubProps, GithubState> {

  constructor(props: GithubProps) {
    super(props);
    this.state = {
    };
  }

  openGithub() {

  }

  render() {
    return (
      <Button basic><Icon name="github" /> Sign in with Github</Button>
    )
  }

}
