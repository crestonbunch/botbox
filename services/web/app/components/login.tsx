import * as React from "react";
import "whatwg-fetch";
import { observer } from "mobx-react";
import { Message, Input, Popup, Button, Form } from "semantic-ui-react";
import { LoginStore } from "../stores/ui/login"

export interface LoginProps {
  trigger: JSX.Element;
  loginStore: LoginStore;
}

@observer
export class Login extends React.Component<LoginProps, {}> {

  render() {
    const loginStore = this.props.loginStore;

    const errMsg = (loginStore.error !== "") ? (
      <Message error>{loginStore.error}</Message>
    ) : null;

    const form = (<Form loading={loginStore.busy}>
      {errMsg}
      <Form.Field>
        <Input icon='mail' iconPosition='left' placeholder="Email"
          onChange={(_, val) => loginStore.email = val.value} />
      </Form.Field>
      <Form.Field>
        <Input type="password" icon='lock' iconPosition='left'
          placeholder="Password"
          onChange={(_, val) => loginStore.password = val.value} />
      </Form.Field>
      <Button primary onClick={(e) => {e.preventDefault(); loginStore.doLogin()}}>
        Login
      </Button>
    </Form>)

    return (<Popup
      trigger={this.props.trigger}
      content={form}
      on='click'
      positioning='bottom center'
    />)
  }

}


