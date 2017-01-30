import * as React from "react";
import { observer } from "mobx-react";
import { Divider, Grid, Image, Message, Icon, Form, Button, Container, Segment } from "semantic-ui-react";
import { Input } from "./components/input";
import { Recaptcha } from "./components/recaptcha";
import { RegisterStore } from "./stores/ui/register";
import { Settings } from "./settings";

export interface RegisterProps {
  registerStore: RegisterStore;
}

export interface RegisterState {
}

@observer
export class Register extends React.Component<RegisterProps, RegisterState> {

  constructor(props: RegisterProps) {
    super(props);
  }

  render() {
    const store = this.props.registerStore;

    const successMsg = (store.success === true) ? (
      <Message success>
        You have been successfully registered!
        Please check your email to finish verifying your account.
      </Message>
    ) : null;

    const errMsg = (store.serverError !== "") ? (
      <Message error>{store.serverError}</Message>
    ) : null;

    const form = (store.success === false) ? (
      <Form error={store.serverError != ""} loading={store.busy}>
        <Input label="Display name"
          placeholder="Enter a display name."
          error={store.nameError}
          onChange={(val) => store.name = val}
          type="text" />
        <Input label="Email"
          placeholder="Enter a valid email."
          error={store.emailError}
          onChange={(val) => store.email = val}
          type="text" />
        <Input label="Password"
          placeholder="Enter a good password."
          error={store.passwordError}
          onChange={(val) => store.password = val}
          type="password" />
        <Recaptcha sitekey={Settings.RECAPTCHA_KEY}
          error={store.recaptchaError}
          onChange={(val) => store.recaptcha = val} />
        <Divider hidden />
        <Button primary
          disabled={store.validating}
          onClick={(e) => { e.preventDefault(); store.doRegister() }}>
          Join
        </Button>
      </Form>
    ) : null;

    return (
      <Segment basic vertical>
        <Image size="medium" src="assets/botbox-masthead.svg" centered />
        <Container>
          <Grid padded centered>
            <Grid.Column largeScreen={5} computer={8} tablet={10} mobile={14}>
              <Segment>
                {successMsg}
                {errMsg}
                {form}
              </Segment>
            </Grid.Column>
          </Grid>
        </Container>
      </Segment>
    )
  }
}