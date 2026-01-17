# bash completion for oci-tf-bootstrap

_oci_tf_bootstrap() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="--profile --config --config-file --output --region --always-free --json --version --help"

    case "${prev}" in
        --profile)
            # Complete with OCI profile names from config
            if [[ -f ~/.oci/config ]]; then
                local profiles=$(grep '^\[' ~/.oci/config | tr -d '[]')
                COMPREPLY=( $(compgen -W "${profiles}" -- ${cur}) )
            fi
            return 0
            ;;
        --config)
            # Complete with directories
            COMPREPLY=( $(compgen -d -- ${cur}) )
            return 0
            ;;
        --config-file)
            # Complete with files
            COMPREPLY=( $(compgen -f -- ${cur}) )
            return 0
            ;;
        --output)
            # Complete with directories
            COMPREPLY=( $(compgen -d -- ${cur}) )
            return 0
            ;;
        --region)
            # Complete with common OCI regions
            local regions="us-ashburn-1 us-phoenix-1 us-sanjose-1 us-chicago-1 eu-frankfurt-1 eu-amsterdam-1 eu-zurich-1 eu-madrid-1 uk-london-1 uk-cardiff-1 ap-tokyo-1 ap-osaka-1 ap-seoul-1 ap-sydney-1 ap-melbourne-1 ap-mumbai-1 ap-hyderabad-1 ca-toronto-1 ca-montreal-1 sa-saopaulo-1 sa-santiago-1 me-jeddah-1 me-dubai-1 af-johannesburg-1"
            COMPREPLY=( $(compgen -W "${regions}" -- ${cur}) )
            return 0
            ;;
        *)
            ;;
    esac

    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}

complete -F _oci_tf_bootstrap oci-tf-bootstrap
