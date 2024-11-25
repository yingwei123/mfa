package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yingwei123/mfa"
)

const codeDuration = 1 * time.Minute
const mfaEmailTemplate = "<div>Your MFA code is {{.MfaCode}}.</div>"
const sendTo = "exampe@example.com"                //the email to send the mfa code to
const clientEmail = "client email for smtp server" //the email of the mfa client
const clientPassword = "password for smtp server"  //the password of the mfa client
const smtpServer = "smtp.gmail.com"                //the smtp server of the mfa client
const smtpServerPort = 587

// example of how to use the mfa client
func main() {
	mfaClient, err := mfa.CreateMFAClient( //create the mfa client
		codeDuration,
		clientEmail,
		clientPassword,
		smtpServer,
		smtpServerPort,
		mfaEmailTemplate,
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = mfaClient.SendMFAEmail(context.Background(), sendTo) //send mfa to yingwei82599@gmail.com
	if err != nil {
		fmt.Println(err.Error())
	}

	input := ""
	var verificationError error

	for { //loop until the mfa code is verified, enter the mfa in the terminal to verify
		fmt.Println("Enter MFA code:")
		fmt.Scanln(&input)

		// Exit the loop if no input is provided
		if input == "" {
			fmt.Println("No input provided. Exiting.")
			break
		}

		verificationError = mfaClient.VerifyMFA(sendTo, input)
		if verificationError == nil {
			fmt.Println("MFA code is verified")
			break
		} else {
			fmt.Println(verificationError.Error() + " Resending MFA code.")
			mfaClient.SendMFAEmail(context.Background(), sendTo)
		}
	}
}
