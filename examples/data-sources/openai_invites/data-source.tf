# List all invites (no filtering arguments supported)
data "openai_invites" "all" {
}

# Output total invites count
output "total_invites" {
  value = length(data.openai_invites.all.invites)
}
