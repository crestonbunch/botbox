import { SessionData } from "./store"

export class Api {

    static MAX_NAME_LENGTH = 20;
    static MIN_PASSWORD_LENGTH = 6;

    /**
     * Authenticate a user and return information about the session.
     */
    static login(email: string, password: string): Promise<SessionData> {
        var session: SessionData = {
            id: 0,
            name: "",
            email: "",
            permissionSet: "",
            permissions: [],
            secret: "",
            expiration: new Date(),
        }

        /**
         * An expected response from the server after authenticating.
         */
        interface SessionPostResponse {
            user: number;
            secret: string;
            expiration: string;
        }

        /**
         * An expected response from the server after getting a user.
         */
        interface UserIdGetResponse {
            name: string;
            joined: Date;
            permission_set: string;
            permissions: string[];
        }

        return fetch('/api/session', {
            method: "POST",
            body: JSON.stringify({ email: email, password: password })
        }).then(function(response: Response) {
            if (response.status == 200) {
                return response.json();
            } else {
                return response.text().then((value: string) => {
                    throw value;
                });
            }
        }).catch(function(reason: any) {
            throw reason;
        }).then(function(value: SessionPostResponse) {
            session.id = value.user;
            session.secret = value.secret;
            session.expiration = new Date(value.expiration);
            return fetch('/api/user/id/' + String(value.user));
        }).catch(function(reason: any) {
            throw reason;
        }).then(function(response: Response) {
            if (response.status == 200) {
                return response.json();
            } else {
                return response.text().then(function(value: string) {
                    throw value;
                });
            }
        }).catch(function(reason: any) {
            throw reason;
        }).then(function(value: UserIdGetResponse) {
            session.name = value.name;
            session.permissions = value.permissions;
            session.permissionSet = value.permission_set;

            return session;
        }).catch(function(reason: any) {
            throw reason;
        });
    }

    static register(
        name: string, email: string, password: string, captcha: string
    ): Promise<void> {

        let request = {
            name: name,
            email: email,
            password: password,
            captcha: captcha,
        }

        // Create a session for the user on the server
        return fetch('/api/user', {
            method: 'POST',
            body: JSON.stringify(request)
        }).then(function(response: Response) {
            if (response.status != 200) {
                // received an error response
                return response.text().then((value: string) => {
                    throw value;
                });
            }
        }).catch(function(reason: any) {
            throw reason;
        }).then(function(value: void) {
            // User account was created, so send an email.
            return fetch('/api/email/verify', {
                method: 'POST',
                body: JSON.stringify({ email: email })
            });
        }).catch(function(reason: any) {
            throw reason;
        }).then(function(response: Response) {
            // There was an error with the email sending service.
            if (response.status != 200) {
                return response.text().then((message: string) => {
                    throw message;
                });
            }
        }).catch(function(reason: any) {
            throw reason;
        });
    }
}