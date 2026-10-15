package cmd

import (
	"github.com/urfave/cli/v3"
)

// UserData represents the data command
var UserData = &cli.Command{
	Name:  "userdata",
	Usage: "ezBookkeeping user data maintenance",
	Commands: []*cli.Command{
		{
			Name:   "user-add",
			Usage:  "Add new user",
			Action: bindAction(addNewUser),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "New user name",
				},
				&cli.StringFlag{
					Name:     "email",
					Aliases:  []string{"m"},
					Required: true,
					Usage:    "New user email",
				},
				&cli.StringFlag{
					Name:     "nickname",
					Aliases:  []string{"i"},
					Required: true,
					Usage:    "New user nickname",
				},
				&cli.StringFlag{
					Name:     "password",
					Aliases:  []string{"p"},
					Required: true,
					Usage:    "New user password",
				},
				&cli.StringFlag{
					Name:     "default-currency",
					Aliases:  []string{"c"},
					Required: true,
					Usage:    "New user default currency",
				},
			},
		},
		{
			Name:   "user-get",
			Usage:  "Get specified user info",
			Action: bindAction(getUserInfo),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-modify-password",
			Usage:  "Modify user password",
			Action: bindAction(modifyUserPassword),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "password",
					Aliases:  []string{"p"},
					Required: true,
					Usage:    "User new password",
				},
			},
		},
		{
			Name:   "user-enable",
			Usage:  "Enable specified user",
			Action: bindAction(enableUser),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-disable",
			Usage:  "Disable specified user",
			Action: bindAction(disableUser),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-set-restrict-features",
			Usage:  "Set restrictions of user features",
			Action: bindAction(setUserFeatureRestriction),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "features",
					Aliases:  []string{"t"},
					Required: true,
					Usage:    "Specific feature types (feature types separated by commas)",
				},
			},
		},
		{
			Name:   "user-add-restrict-features",
			Usage:  "Add restrictions of user features",
			Action: bindAction(addUserFeatureRestriction),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "features",
					Aliases:  []string{"t"},
					Required: true,
					Usage:    "Specific feature types (feature types separated by commas)",
				},
			},
		},
		{
			Name:   "user-remove-restrict-features",
			Usage:  "Remove restrictions of user features",
			Action: bindAction(removeUserFeatureRestriction),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "features",
					Aliases:  []string{"t"},
					Required: true,
					Usage:    "Specific feature types (feature types separated by commas)",
				},
			},
		},
		{
			Name:   "user-resend-verify-email",
			Usage:  "Resend user verify email",
			Action: bindAction(resendUserVerifyEmail),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-set-email-verified",
			Usage:  "Set user email address verified",
			Action: bindAction(setUserEmailVerified),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-set-email-unverified",
			Usage:  "Set user email address unverified",
			Action: bindAction(setUserEmailUnverified),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-delete",
			Usage:  "Delete specified user",
			Action: bindAction(deleteUser),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-2fa-disable",
			Usage:  "Disable user 2fa setting",
			Action: bindAction(disableUser2FA),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-session-list",
			Usage:  "List all user sessions",
			Action: bindAction(listUserTokens),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "user-session-new",
			Usage:  "Create new session for user",
			Action: bindAction(createNewUserToken),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "type",
					Aliases:  []string{"t"},
					Required: false,
					Usage:    "Specific token type, supports \"api\" and \"mcp\", default is \"api\"",
				},
				&cli.Int64Flag{
					Name:     "expiresInSeconds",
					Aliases:  []string{"e"},
					Required: true,
					Usage:    "Token expiration time in seconds (0 - 4294967295, 0 means no expiration).",
				},
			},
		},
		{
			Name:   "user-session-revoke",
			Usage:  "Revoke the specified user session",
			Action: bindAction(revokeUserToken),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "token",
					Aliases:  []string{"t"},
					Required: false,
					Usage:    "Specific token content",
				},
			},
		},
		{
			Name:   "user-session-clear",
			Usage:  "Clear user all sessions",
			Action: bindAction(clearUserTokens),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "send-password-reset-mail",
			Usage:  "Send password reset mail",
			Action: bindAction(sendPasswordResetMail),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "transaction-check",
			Usage:  "Check whether user all transactions and accounts are correct",
			Action: bindAction(checkUserTransactionAndAccount),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "transaction-tag-index-fix-transaction-time",
			Usage:  "Fix the transaction tag index data which does not have transaction time",
			Action: bindAction(fixTransactionTagIndexNotHaveTransactionTime),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
			},
		},
		{
			Name:   "transaction-import",
			Usage:  "Import transactions to specified user",
			Action: bindAction(importUserTransaction),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "file",
					Aliases:  []string{"f"},
					Required: true,
					Usage:    "Specific import file path (e.g. transaction.csv)",
				},
				&cli.StringFlag{
					Name:     "type",
					Aliases:  []string{"t"},
					Required: true,
					Usage:    "Import file type (supports \"ezbookkeeping_csv\", \"ezbookkeeping_tsv\")",
				},
			},
		},
		{
			Name:   "transaction-export",
			Usage:  "Export user all transactions to file",
			Action: bindAction(exportUserTransaction),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"n"},
					Required: true,
					Usage:    "Specific user name",
				},
				&cli.StringFlag{
					Name:     "file",
					Aliases:  []string{"f"},
					Required: true,
					Usage:    "Specific exported file path (e.g. transaction.csv)",
				},
				&cli.StringFlag{
					Name:     "type",
					Aliases:  []string{"t"},
					Required: false,
					Usage:    "Export file type, support csv or tsv, default is csv",
				},
			},
		},
	},
}
