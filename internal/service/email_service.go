package service

import (
	"bytes"
	"drivo/internal/app"
	"drivo/internal/jobs"
	"fmt"
	"html/template"
	"time"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

type MailService struct {
	cfg *app.App
}

func NewMailService(cfg *app.App) *MailService {
	return &MailService{cfg: cfg}
}

type EmailType string

const (
	EmailTypeOTP              EmailType = "otp"
	EmailTypeWelcome          EmailType = "welcome"
	EmailTypeRideConfirmation EmailType = "ride_confirmation"
	EmailTypeRideCompleted    EmailType = "ride_completed"
	EmailTypeDriverApproved   EmailType = "driver_approved"
)

type OTPEmailData struct {
	Name      string
	OTP       string
	ExpiresIn string
	Year      int
}

type WelcomeEmailData struct {
	Name string
	Year int
}

func (s *MailService) SendOTPEmail(to, name, otp string) error {
	data := OTPEmailData{
		Name:      name,
		OTP:       otp,
		ExpiresIn: "10 minutes",
		Year:      time.Now().Year(),
	}

	html, err := renderTemplate(otpTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render OTP template: %v", err)
	}

	return s.send(to, "Verify Your Drivo Account", html)
}

func (s *MailService) SendWelcomeEmail(to, name string) error {
	data := WelcomeEmailData{
		Name: name,
		Year: time.Now().Year(),
	}

	html, err := renderTemplate(welcomeTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render welcome template: %v", err)
	}

	return s.send(to, "Welcome to Drivo!", html)
}

func (s *MailService) SendRideConfirmationEmail(to string, data jobs.RideConfirmationData) error {
	data.Year = time.Now().Year()

	html, err := renderTemplate(rideConfirmationTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render ride confirmation template: %v", err)
	}

	return s.send(to, "Your Drivo Ride is Confirmed", html)
}

func (s *MailService) SendDriverWelcomeEmail(to, name string) error {
	data := WelcomeEmailData{
		Name: name,
		Year: time.Now().Year(),
	}

	html, err := renderTemplate(driverWelcomeTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render driver welcome template: %v", err)
	}

	return s.send(to, "Welcome to Drivo — Let's Get You on the Road", html)
}

func (s *MailService) SendRideCompletedEmail(to string, data jobs.RideCompletedData) error {
	data.Year = time.Now().Year()

	html, err := renderTemplate(rideCompletedTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render ride completed template: %v", err)
	}

	return s.send(to, "Your Drivo Ride Receipt", html)
}

func (s *MailService) SendDriverApprovedEmail(to string, name string) error {
	data := WelcomeEmailData{
		Name: name,
		Year: time.Now().Year(),
	}

	html, err := renderTemplate(driverApprovedTemplate, data)
	if err != nil {
		return fmt.Errorf("failed to render driver approved template: %v", err)
	}

	return s.send(to, "Your Drivo Driver Account is Approved 🎉", html)
}

func (s *MailService) send(to, subject, htmlBody string) error {
	client := mailjet.NewMailjetClient(
		s.cfg.Config.MailjetAPIKey,
		s.cfg.Config.MailjetSecretKey,
	)

	messages := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: s.cfg.Config.MailjetFromEmail,
					Name:  s.cfg.Config.MailjetFromName,
				},
				To: &mailjet.RecipientsV31{
					{Email: to},
				},
				Subject:  subject,
				HTMLPart: htmlBody,
			},
		},
	}

	_, err := client.SendMailV31(&messages)
	if err != nil {
		return fmt.Errorf("mailjet error: %v", err)
	}

	return nil
}

func renderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const driverApprovedTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #fff; font-size: 28px; font-weight: 700; }
  .header span { color: #facc15; }
  .hero { background: #facc15; padding: 40px; text-align: center; }
  .hero-icon { font-size: 56px; margin-bottom: 12px; }
  .hero h2 { font-size: 26px; font-weight: 700; color: #000; margin-bottom: 8px; }
  .hero p { font-size: 15px; color: #000; opacity: 0.7; }
  .body { padding: 40px; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .checklist { margin: 24px 0; }
  .check-item { display: flex; align-items: center; gap: 12px; padding: 12px 0; border-bottom: 1px solid #f3f4f6; }
  .check-item:last-child { border-bottom: none; }
  .check-icon { width: 24px; height: 24px; background: #000; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; color: #facc15; font-size: 13px; font-weight: 700; text-align: center; line-height: 24px; }
  .check-text { font-size: 14px; color: #374151; font-weight: 500; }
  .divider { border: none; border-top: 1px solid #e5e7eb; margin: 28px 0; }
  .stats { display: flex; gap: 12px; margin: 24px 0; }
  .stat { flex: 1; background: #f9fafb; border-radius: 10px; padding: 16px 12px; text-align: center; border: 1px solid #e5e7eb; }
  .stat-value { font-size: 20px; font-weight: 700; color: #000; }
  .stat-label { font-size: 11px; color: #6b7280; margin-top: 4px; text-transform: uppercase; letter-spacing: 0.5px; }
  .cta-box { background: #000; border-radius: 12px; padding: 28px; text-align: center; margin: 28px 0; }
  .cta-box p { color: #9ca3af; font-size: 14px; margin-bottom: 16px; }
  .cta-box h3 { color: #fff; font-size: 18px; font-weight: 600; margin-bottom: 8px; }
  .tip { background: #fffbeb; border: 1px solid #fde68a; border-radius: 10px; padding: 16px 20px; margin: 20px 0; }
  .tip p { font-size: 14px; color: #92400e; margin: 0; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; line-height: 1.6; }
</style>
</head>
<body>
<div class="wrapper">

  <div class="header">
    <h1>Driv<span>o</span></h1>
  </div>

  <div class="hero">
    <div class="hero-icon">🎉</div>
    <h2>You're Approved!</h2>
    <p>Your driver account is fully verified and ready to go</p>
  </div>

  <div class="body">

    <p>Hi <strong>{{.Name}}</strong>,</p>
    <p>Great news — our team has reviewed your documents and your Drivo driver account has been fully approved. You can now go online and start accepting rides.</p>

    <div class="checklist">
      <div class="check-item">
        <div class="check-icon">✓</div>
        <div class="check-text">Identity verified</div>
      </div>
      <div class="check-item">
        <div class="check-icon">✓</div>
        <div class="check-text">Driver's license verified</div>
      </div>
      <div class="check-item">
        <div class="check-icon">✓</div>
        <div class="check-text">Vehicle verified</div>
      </div>
      <div class="check-item">
        <div class="check-icon">✓</div>
        <div class="check-text">Account activated</div>
      </div>
    </div>

    <div class="stats">
      <div class="stat">
        <div class="stat-value">₦0</div>
        <div class="stat-label">Earnings</div>
      </div>
      <div class="stat">
        <div class="stat-value">5.0 ⭐</div>
        <div class="stat-label">Rating</div>
      </div>
      <div class="stat">
        <div class="stat-value">0</div>
        <div class="stat-label">Trips</div>
      </div>
    </div>

    <div class="cta-box">
      <h3>Ready to start earning?</h3>
      <p>Open the Drivo app, go online, and your first ride request will come in shortly.</p>
    </div>

    <hr class="divider">

    <p><strong>Quick reminders before your first trip:</strong></p>

    <div class="tip">
      <p>⏱ You have <strong>15 seconds</strong> to accept each ride request. Missing requests lowers your acceptance rate.</p>
    </div>

    <div class="tip">
      <p>⭐ Drivers with a rating above <strong>4.8</strong> get priority in ride matching. Be punctual, polite, and keep your vehicle clean.</p>
    </div>

    <div class="tip">
      <p>📍 Always make sure your location is enabled so riders can find you accurately.</p>
    </div>

    <hr class="divider">

    <p>If you have any questions or need help getting started, our support team is available anytime.</p>
    <p>Welcome to the road,<br><strong>The Drivo Team</strong></p>

  </div>

  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.<br>
    You received this email because your Drivo driver account was approved.</p>
  </div>

</div>
</body>
</html>
`

const baseLayout = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; color: #111827; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #ffffff; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
  .header span { color: #facc15; }
  .body { padding: 40px; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; line-height: 1.6; }
  h2 { font-size: 22px; font-weight: 600; margin-bottom: 12px; color: #111827; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .otp-box { background: #f9fafb; border: 2px dashed #d1d5db; border-radius: 10px; text-align: center; padding: 24px; margin: 24px 0; }
  .otp-code { font-size: 40px; font-weight: 700; letter-spacing: 12px; color: #000000; }
  .otp-hint { font-size: 13px; color: #6b7280; margin-top: 8px; }
  .info-card { background: #f9fafb; border-radius: 10px; padding: 20px 24px; margin: 20px 0; }
  .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e5e7eb; font-size: 14px; }
  .info-row:last-child { border-bottom: none; }
  .info-label { color: #6b7280; }
  .info-value { font-weight: 600; color: #111827; }
  .badge { display: inline-block; background: #000000; color: #ffffff; padding: 4px 12px; border-radius: 20px; font-size: 13px; font-weight: 500; }
  .fare { font-size: 32px; font-weight: 700; color: #000000; text-align: center; margin: 20px 0; }
  .divider { border: none; border-top: 1px solid #e5e7eb; margin: 24px 0; }
</style>
</head>
<body>
<div class="wrapper">
  <div class="header">
    <h1>Driv<span>o</span></h1>
  </div>
  <div class="body">
    {{.Content}}
  </div>
  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.<br>You received this email because you have a Drivo account.</p>
  </div>
</div>
</body>
</html>
`

const otpTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #fff; font-size: 28px; font-weight: 700; letter-spacing: -0.5px; }
  .header span { color: #facc15; }
  .body { padding: 40px; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; line-height: 1.6; }
  h2 { font-size: 22px; font-weight: 600; margin-bottom: 12px; color: #111827; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .otp-box { background: #f9fafb; border: 2px dashed #d1d5db; border-radius: 10px; text-align: center; padding: 24px; margin: 24px 0; }
  .otp-code { font-size: 40px; font-weight: 700; letter-spacing: 12px; color: #000; }
  .otp-hint { font-size: 13px; color: #6b7280; margin-top: 8px; }
  .warning { font-size: 13px; color: #6b7280; }
</style>
</head>
<body>
<div class="wrapper">
  <div class="header"><h1>Driv<span>o</span></h1></div>
  <div class="body">
    <h2>Verify your account</h2>
    <p>Hi {{.Name}},</p>
    <p>Use the code below to verify your Drivo account. This code expires in <strong>{{.ExpiresIn}}</strong>.</p>
    <div class="otp-box">
      <div class="otp-code">{{.OTP}}</div>
      <div class="otp-hint">Enter this code in the app to continue</div>
    </div>
    <p class="warning">If you did not request this, you can safely ignore this email.</p>
  </div>
  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.</p>
  </div>
</div>
</body>
</html>
`

const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #fff; font-size: 28px; font-weight: 700; }
  .header span { color: #facc15; }
  .body { padding: 40px; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; line-height: 1.6; }
  h2 { font-size: 22px; font-weight: 600; margin-bottom: 12px; color: #111827; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .highlight { background: #f9fafb; border-left: 4px solid #000; padding: 16px 20px; border-radius: 0 8px 8px 0; margin: 20px 0; }
  .cta { display: block; background: #000; color: #fff; text-align: center; padding: 14px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 15px; margin: 24px 0; }
</style>
</head>
<body>
<div class="wrapper">
  <div class="header"><h1>Driv<span>o</span></h1></div>
  <div class="body">
    <h2>Welcome to Drivo, {{.Name}}! 🎉</h2>
    <p>Your account has been verified and you're all set to start riding.</p>
    <div class="highlight">
      <p style="margin:0; font-weight: 600;">Here's what you can do now:</p>
      <p style="margin: 8px 0 0 0;">Request rides, track your driver in real time, and pay seamlessly — all from the Drivo app.</p>
    </div>
    <p>If you have any questions, our support team is always here to help.</p>
    <p>Safe travels,<br><strong>The Drivo Team</strong></p>
  </div>
  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.</p>
  </div>
</div>
</body>
</html>
`

const rideConfirmationTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #fff; font-size: 28px; font-weight: 700; }
  .header span { color: #facc15; }
  .body { padding: 40px; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; }
  h2 { font-size: 22px; font-weight: 600; margin-bottom: 12px; color: #111827; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .info-card { background: #f9fafb; border-radius: 10px; padding: 20px 24px; margin: 20px 0; }
  .info-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e5e7eb; font-size: 14px; }
  .info-row:last-child { border-bottom: none; }
  .info-label { color: #6b7280; }
  .info-value { font-weight: 600; color: #111827; }
  .route { background: #000; color: #fff; border-radius: 10px; padding: 20px 24px; margin: 20px 0; }
  .route-point { display: flex; align-items: flex-start; gap: 12px; padding: 8px 0; }
  .dot { width: 10px; height: 10px; border-radius: 50%; margin-top: 5px; flex-shrink: 0; }
  .dot-green { background: #4ade80; }
  .dot-red { background: #f87171; }
  .route-label { font-size: 12px; color: #9ca3af; }
  .route-value { font-size: 14px; font-weight: 500; color: #fff; }
</style>
</head>
<body>
<div class="wrapper">
  <div class="header"><h1>Driv<span>o</span></h1></div>
  <div class="body">
    <h2>Your ride is confirmed </h2>
    <p>Hi {{.RiderName}}, driver accepted!</p>

    <div class="route">
      <div class="route-point">
        <div class="dot dot-green"></div>
        <div>
          <div class="route-label">PICKUP</div>
          <div class="route-value">{{.PickupAddress}}</div>
        </div>
      </div>
      <div class="route-point">
        <div class="dot dot-red"></div>
        <div>
          <div class="route-label">DROPOFF</div>
          <div class="route-value">{{.DropoffAddress}}</div>
        </div>
      </div>
    </div>

    <div class="info-card">
      <div class="info-row">
        <span class="info-label">Driver</span>
        <span class="info-value">{{.DriverName}}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Vehicle</span>
        <span class="info-value">{{.VehicleColor}} {{.VehicleMake}} {{.VehicleModel}}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Plate Number</span>
        <span class="info-value">{{.PlateNumber}}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Estimated Fare</span>
        <span class="info-value">₦{{.EstimatedFare}}</span>
      </div>
      <div class="info-row">
        <span class="info-label">ETA</span>
        <span class="info-value">{{.ETA}} minutes</span>
      </div>
    </div>
  </div>
  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.</p>
  </div>
</div>
</body>
</html>
`

const driverWelcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #fff; font-size: 28px; font-weight: 700; }
  .header span { color: #facc15; }
  .banner { background: #facc15; padding: 20px 40px; text-align: center; }
  .banner p { font-size: 15px; font-weight: 600; color: #000; }
  .body { padding: 40px; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; line-height: 1.6; }
  h2 { font-size: 22px; font-weight: 600; margin-bottom: 12px; color: #111827; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .steps { margin: 24px 0; }
  .step { display: flex; gap: 16px; margin-bottom: 20px; align-items: flex-start; }
  .step-number { background: #000; color: #facc15; width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 14px; flex-shrink: 0; text-align: center; line-height: 32px; }
  .step-content h3 { font-size: 15px; font-weight: 600; color: #111827; margin-bottom: 4px; }
  .step-content p { font-size: 14px; color: #6b7280; margin: 0; }
  .highlight { background: #f9fafb; border-left: 4px solid #facc15; padding: 16px 20px; border-radius: 0 8px 8px 0; margin: 24px 0; }
  .highlight p { margin: 0; font-size: 14px; color: #374151; }
  .divider { border: none; border-top: 1px solid #e5e7eb; margin: 24px 0; }
  .stats { display: flex; gap: 16px; margin: 24px 0; }
  .stat { flex: 1; background: #f9fafb; border-radius: 10px; padding: 16px; text-align: center; }
  .stat-value { font-size: 22px; font-weight: 700; color: #000; }
  .stat-label { font-size: 12px; color: #6b7280; margin-top: 4px; }
</style>
</head>
<body>
<div class="wrapper">

  <div class="header">
    <h1>Driv<span>o</span></h1>
  </div>

  <div class="banner">
    <p>🎉 You're officially a Drivo Driver</p>
  </div>

  <div class="body">
    <h2>Welcome aboard, {{.Name}}!</h2>
    <p>Your application has been verified and your driver account is now active. You're ready to start earning with Drivo.</p>

    <div class="stats">
      <div class="stat">
        <div class="stat-value">₦0</div>
        <div class="stat-label">Earnings so far</div>
      </div>
      <div class="stat">
        <div class="stat-value">5.0 ⭐</div>
        <div class="stat-label">Starting rating</div>
      </div>
      <div class="stat">
        <div class="stat-value">0</div>
        <div class="stat-label">Trips completed</div>
      </div>
    </div>

    <hr class="divider">

    <p><strong>Here's how to get started:</strong></p>

    <div class="steps">
      <div class="step">
        <div class="step-number">1</div>
        <div class="step-content">
          <h3>Open the Drivo Driver App</h3>
          <p>Log in with your registered email and password.</p>
        </div>
      </div>
      <div class="step">
        <div class="step-number">2</div>
        <div class="step-content">
          <h3>Go Online</h3>
          <p>Tap the "Go Online" button to start receiving ride requests in your area.</p>
        </div>
      </div>
      <div class="step">
        <div class="step-number">3</div>
        <div class="step-content">
          <h3>Accept Ride Requests</h3>
          <p>You have 15 seconds to accept each request. Keep your acceptance rate high for better opportunities.</p>
        </div>
      </div>
      <div class="step">
        <div class="step-number">4</div>
        <div class="step-content">
          <h3>Complete Trips &amp; Earn</h3>
          <p>Pick up riders, complete trips, and watch your earnings grow.</p>
        </div>
      </div>
    </div>

    <div class="highlight">
      <p>💡 <strong>Pro tip:</strong> Drivers with a rating above 4.8 get priority placement in ride requests. Always be on time, keep your vehicle clean, and be courteous to riders.</p>
    </div>

    <hr class="divider">

    <p>If you have any questions or need support, reach out to us anytime. We're excited to have you on the Drivo team.</p>
    <p>Drive safe,<br><strong>The Drivo Team</strong></p>
  </div>

  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.<br>
    You received this email because you registered as a driver on Drivo.</p>
  </div>

</div>
</body>
</html>
`

const rideCompletedTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f4f4f5; }
  .wrapper { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .header { background: #000; padding: 32px 40px; text-align: center; }
  .header h1 { color: #fff; font-size: 28px; font-weight: 700; }
  .header span { color: #facc15; }
  .body { padding: 40px; }
  .footer { background: #f9fafb; padding: 24px 40px; text-align: center; border-top: 1px solid #e5e7eb; }
  .footer p { color: #9ca3af; font-size: 13px; }
  h2 { font-size: 22px; font-weight: 600; margin-bottom: 12px; color: #111827; }
  p { font-size: 15px; line-height: 1.7; color: #374151; margin-bottom: 16px; }
  .fare-box { text-align: center; padding: 32px; background: #000; border-radius: 12px; margin: 24px 0; }
  .fare-label { color: #9ca3af; font-size: 13px; margin-bottom: 8px; }
  .fare-amount { font-size: 48px; font-weight: 700; color: #facc15; }
  .info-card { background: #f9fafb; border-radius: 10px; padding: 20px 24px; margin: 20px 0; }
  .info-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e5e7eb; font-size: 14px; }
  .info-row:last-child { border-bottom: none; }
  .info-label { color: #6b7280; }
  .info-value { font-weight: 600; color: #111827; }
</style>
</head>
<body>
<div class="wrapper">
  <div class="header"><h1>Driv<span>o</span></h1></div>
  <div class="body">
    <h2>Trip completed 🏁</h2>
    <p>Hi {{.RiderName}}, thanks for riding with Drivo. Here's your receipt.</p>

    <div class="fare-box">
      <div class="fare-label">TOTAL FARE</div>
      <div class="fare-amount">₦{{.ActualFare}}</div>
    </div>

    <div class="info-card">
      <div class="info-row">
        <span class="info-label">Pickup</span>
        <span class="info-value">{{.PickupAddress}}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Dropoff</span>
        <span class="info-value">{{.DropoffAddress}}</span>
      </div>
      <div class="info-row">
        <span class="info-label">Distance</span>
        <span class="info-value">{{.DistanceKm}} km</span>
      </div>
    </div>

    <p>We hope you had a great ride. See you next time!</p>
  </div>
  <div class="footer">
    <p>© {{.Year}} Drivo. All rights reserved.</p>
  </div>
</div>
</body>
</html>
`
