import * as React from "react";
import "whatwg-fetch"
import { Label, Message, Input, Icon, Form, Button } from "semantic-ui-react";

const PASSWORD_LENGTH = 6;

export interface PasswordProps {
}

export interface PasswordState {
  error?: string
  value?: string
}

export class Password extends React.Component<PasswordProps, PasswordState> {

  constructor(props: PasswordProps) {
    super(props);
    this.state = {
      value: "",
    };
  }

  getValue(): string {
    return this.state.value;
  }

  updatePassword(e: Event) {
    let value = (e.target as any).value;

    this.setState({
      value: value,
      error: undefined
    });
  }

  validate(): Promise<boolean> {
    if (this.state.value == "") {
      this.setState({
        error: "Please provide a password."
      });
      return new Promise<boolean>(function(r) {r(false);});
    } else if (this.state.value.length < PASSWORD_LENGTH) {
      this.setState({
        error: "Password must be at least 6 characters."
      });
      return new Promise<boolean>(function(r) {r(false);});
    }

    return new Promise<boolean>(function(r) {r(true);});
  }

  render() {
    return (
    <Form.Field required error={this.state.error != undefined}>
      <label>Password</label>
      {this.state.error != undefined ?
        <Label basic pointing="below" color="red"><Icon name="warning sign" />
          {this.state.error}
        </Label>
      : null}
      <Input icon='lock' iconPosition='left'
        onChange={this.updatePassword.bind(this)} type="password"
        name="password" placeholder="Make it secure" />
    </Form.Field>
    )
  }

}
