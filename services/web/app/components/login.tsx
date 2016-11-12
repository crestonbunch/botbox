import * as React from "react";
import "whatwg-fetch"
import { Input, Popup, Button, Form } from "semantic-ui-react";

export interface LoginProps {
  trigger: JSX.Element;
}

export class Login extends React.Component<LoginProps, {}> {

  render() {

    const form = (<Form>
      <Form.Field>
        <label>Username</label>
        <Input icon='users' iconPosition='left' />
      </Form.Field>
      <Form.Field>
        <label>Password</label>
        <Input type="password" icon='lock' iconPosition='left' />
      </Form.Field>
      <Button primary>Login</Button>
    </Form>)

    return (<Popup
      trigger={this.props.trigger}
      content={form}
      on='click'
      positioning='bottom center'
    />)
  }

}


