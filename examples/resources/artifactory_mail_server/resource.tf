resource "artifactory_mail_server" "mymailserver" {
    enabled         = true
    artifactory_url = "http://tempurl.org"
    from            = "test@jfrog.com"
    host            = "http://tempurl.org"
    username        = "test-user"
    password        = "test-password"
    port            = 25
    subject_prefix  = "[Test]"
    use_ssl         = true
    use_tls         = true
}