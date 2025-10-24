package templates

// Built-in email templates

const welcomeHTMLTemplate = `
{{define "subject"}}Welcome to Video Converter!{{end}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to Video Converter</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Video Converter!</h1>
        </div>
        <div class="content">
            <h2>Hello {{.FirstName}}!</h2>
            <p>Thank you for joining Video Converter. We're excited to help you convert your videos to MP3 format quickly and easily.</p>
            
            <h3>Getting Started:</h3>
            <ul>
                <li>Upload your video files through our web interface</li>
                <li>Monitor conversion progress in real-time</li>
                <li>Download your converted MP3 files</li>
                <li>Access your conversion history anytime</li>
            </ul>
            
            <p>
                <a href="{{.DashboardURL}}" class="button">Go to Dashboard</a>
            </p>
            
            <p>If you have any questions, feel free to contact our support team.</p>
        </div>
        <div class="footer">
            <p>© 2024 Video Converter. All rights reserved.</p>
            <p><a href="{{.UnsubscribeURL}}">Unsubscribe</a> from these emails</p>
        </div>
    </div>
</body>
</html>
`

const welcomeTextTemplate = `
{{define "subject"}}Welcome to Video Converter!{{end}}
Hello {{.FirstName}}!

Thank you for joining Video Converter. We're excited to help you convert your videos to MP3 format quickly and easily.

Getting Started:
- Upload your video files through our web interface
- Monitor conversion progress in real-time
- Download your converted MP3 files
- Access your conversion history anytime

Visit your dashboard: {{.DashboardURL}}

If you have any questions, feel free to contact our support team.

© 2024 Video Converter. All rights reserved.
Unsubscribe: {{.UnsubscribeURL}}
`

const conversionCompleteHTMLTemplate = `
{{define "subject"}}Your video conversion is complete - {{.VideoName}}{{end}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Conversion Complete</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 5px; }
        .success { color: #4CAF50; font-weight: bold; }
        .details { background-color: white; padding: 15px; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Conversion Complete!</h1>
        </div>
        <div class="content">
            <h2>Hello {{.FirstName}}!</h2>
            <p class="success">Your video has been successfully converted to MP3!</p>
            
            <div class="details">
                <h3>Conversion Details:</h3>
                <ul>
                    <li><strong>Original File:</strong> {{.VideoName}}</li>
                    <li><strong>Duration:</strong> {{.Duration}}</li>
                    <li><strong>File Size:</strong> {{.FileSize}}</li>
                    <li><strong>Conversion Time:</strong> {{.ConversionTime}}</li>
                    <li><strong>Quality:</strong> {{.Quality}}</li>
                </ul>
            </div>
            
            <p>Your MP3 file is ready for download:</p>
            <p>
                <a href="{{.DownloadURL}}" class="button">Download MP3</a>
                <a href="{{.DashboardURL}}" class="button">View Dashboard</a>
            </p>
            
            <p><small>Note: Download links expire after 30 days.</small></p>
        </div>
        <div class="footer">
            <p>© 2024 Video Converter. All rights reserved.</p>
            <p><a href="{{.UnsubscribeURL}}">Unsubscribe</a> from these emails</p>
        </div>
    </div>
</body>
</html>
`

const conversionCompleteTextTemplate = `
{{define "subject"}}Your video conversion is complete - {{.VideoName}}{{end}}
Hello {{.FirstName}}!

✅ Your video has been successfully converted to MP3!

Conversion Details:
- Original File: {{.VideoName}}
- Duration: {{.Duration}}
- File Size: {{.FileSize}}
- Conversion Time: {{.ConversionTime}}
- Quality: {{.Quality}}

Your MP3 file is ready for download:
{{.DownloadURL}}

View your dashboard: {{.DashboardURL}}

Note: Download links expire after 30 days.

© 2024 Video Converter. All rights reserved.
Unsubscribe: {{.UnsubscribeURL}}
`

const conversionErrorHTMLTemplate = `
{{define "subject"}}Video conversion failed - {{.VideoName}}{{end}}
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Conversion Failed</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f44336; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 5px; }
        .error { color: #f44336; font-weight: bold; }
        .details { background-color: white; padding: 15px; border-radius: 5px; margin: 10px 0; }
        .troubleshooting { background-color: #fff3cd; padding: 15px; border-radius: 5px; margin: 10px 0; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>❌ Conversion Failed</h1>
        </div>
        <div class="content">
            <h2>Hello {{.FirstName}}!</h2>
            <p class="error">Unfortunately, we encountered an error while converting your video.</p>
            
            <div class="details">
                <h3>Error Details:</h3>
                <ul>
                    <li><strong>File:</strong> {{.VideoName}}</li>
                    <li><strong>Error:</strong> {{.ErrorMessage}}</li>
                    <li><strong>Time:</strong> {{.ErrorTime}}</li>
                    <li><strong>Job ID:</strong> {{.JobID}}</li>
                </ul>
            </div>
            
            <div class="troubleshooting">
                <h3>🔧 Troubleshooting Tips:</h3>
                <ul>
                    <li>Ensure your video file is in a supported format (MP4, AVI, MOV, MKV, WEBM)</li>
                    <li>Check that the file size is under our 500MB limit</li>
                    <li>Verify the video file is not corrupted</li>
                    <li>Try uploading the file again</li>
                </ul>
            </div>
            
            <p>
                <a href="{{.UploadURL}}" class="button">Try Again</a>
                <a href="{{.SupportURL}}" class="button">Contact Support</a>
            </p>
            
            <p>If the problem persists, please contact our support team with the Job ID above.</p>
        </div>
        <div class="footer">
            <p>© 2024 Video Converter. All rights reserved.</p>
            <p><a href="{{.UnsubscribeURL}}">Unsubscribe</a> from these emails</p>
        </div>
    </div>
</body>
</html>
`

const conversionErrorTextTemplate = `
{{define "subject"}}Video conversion failed - {{.VideoName}}{{end}}
Hello {{.FirstName}}!

❌ Unfortunately, we encountered an error while converting your video.

Error Details:
- File: {{.VideoName}}
- Error: {{.ErrorMessage}}
- Time: {{.ErrorTime}}
- Job ID: {{.JobID}}

🔧 Troubleshooting Tips:
- Ensure your video file is in a supported format (MP4, AVI, MOV, MKV, WEBM)
- Check that the file size is under our 500MB limit
- Verify the video file is not corrupted
- Try uploading the file again

Try again: {{.UploadURL}}
Contact support: {{.SupportURL}}

If the problem persists, please contact our support team with the Job ID above.

© 2024 Video Converter. All rights reserved.
Unsubscribe: {{.UnsubscribeURL}}
`